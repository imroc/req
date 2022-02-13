// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package req

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2/hpack"
	"io"
	"log"
	"strings"
	"sync"
)

const http2frameHeaderLen = 9

var http2padZeros = make([]byte, 255) // zeros for padding

// A FrameType is a registered frame type as defined in
// http://http2.github.io/http2-spec/#rfc.section.11.2
type http2FrameType uint8

const (
	http2FrameData         http2FrameType = 0x0
	http2FrameHeaders      http2FrameType = 0x1
	http2FramePriority     http2FrameType = 0x2
	http2FrameRSTStream    http2FrameType = 0x3
	http2FrameSettings     http2FrameType = 0x4
	http2FramePushPromise  http2FrameType = 0x5
	http2FramePing         http2FrameType = 0x6
	http2FrameGoAway       http2FrameType = 0x7
	http2FrameWindowUpdate http2FrameType = 0x8
	http2FrameContinuation http2FrameType = 0x9
)

var http2frameName = map[http2FrameType]string{
	http2FrameData:         "DATA",
	http2FrameHeaders:      "HEADERS",
	http2FramePriority:     "PRIORITY",
	http2FrameRSTStream:    "RST_STREAM",
	http2FrameSettings:     "SETTINGS",
	http2FramePushPromise:  "PUSH_PROMISE",
	http2FramePing:         "PING",
	http2FrameGoAway:       "GOAWAY",
	http2FrameWindowUpdate: "WINDOW_UPDATE",
	http2FrameContinuation: "CONTINUATION",
}

func (t http2FrameType) String() string {
	if s, ok := http2frameName[t]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN_FRAME_TYPE_%d", uint8(t))
}

// Flags is a bitmask of HTTP/2 flags.
// The meaning of flags varies depending on the frame type.
type http2Flags uint8

// Has reports whether f contains all (0 or more) flags in v.
func (f http2Flags) Has(v http2Flags) bool {
	return (f & v) == v
}

// Frame-specific FrameHeader flag bits.
const (
	// Data Frame
	http2FlagDataEndStream http2Flags = 0x1
	http2FlagDataPadded    http2Flags = 0x8

	// Headers Frame
	http2FlagHeadersEndStream  http2Flags = 0x1
	http2FlagHeadersEndHeaders http2Flags = 0x4
	http2FlagHeadersPadded     http2Flags = 0x8
	http2FlagHeadersPriority   http2Flags = 0x20

	// Settings Frame
	http2FlagSettingsAck http2Flags = 0x1

	// Ping Frame
	http2FlagPingAck http2Flags = 0x1

	// Continuation Frame
	http2FlagContinuationEndHeaders http2Flags = 0x4

	http2FlagPushPromiseEndHeaders http2Flags = 0x4
	http2FlagPushPromisePadded     http2Flags = 0x8
)

var http2flagName = map[http2FrameType]map[http2Flags]string{
	http2FrameData: {
		http2FlagDataEndStream: "END_STREAM",
		http2FlagDataPadded:    "PADDED",
	},
	http2FrameHeaders: {
		http2FlagHeadersEndStream:  "END_STREAM",
		http2FlagHeadersEndHeaders: "END_HEADERS",
		http2FlagHeadersPadded:     "PADDED",
		http2FlagHeadersPriority:   "PRIORITY",
	},
	http2FrameSettings: {
		http2FlagSettingsAck: "ACK",
	},
	http2FramePing: {
		http2FlagPingAck: "ACK",
	},
	http2FrameContinuation: {
		http2FlagContinuationEndHeaders: "END_HEADERS",
	},
	http2FramePushPromise: {
		http2FlagPushPromiseEndHeaders: "END_HEADERS",
		http2FlagPushPromisePadded:     "PADDED",
	},
}

// a frameParser parses a frame given its FrameHeader and payload
// bytes. The length of payload will always equal fh.Length (which
// might be 0).
type http2frameParser func(fc *http2frameCache, fh http2FrameHeader, countError func(string), payload []byte) (http2Frame, error)

var http2frameParsers = map[http2FrameType]http2frameParser{
	http2FrameData:         http2parseDataFrame,
	http2FrameHeaders:      http2parseHeadersFrame,
	http2FramePriority:     http2parsePriorityFrame,
	http2FrameRSTStream:    http2parseRSTStreamFrame,
	http2FrameSettings:     http2parseSettingsFrame,
	http2FramePushPromise:  http2parsePushPromise,
	http2FramePing:         http2parsePingFrame,
	http2FrameGoAway:       http2parseGoAwayFrame,
	http2FrameWindowUpdate: http2parseWindowUpdateFrame,
	http2FrameContinuation: http2parseContinuationFrame,
}

func http2typeFrameParser(t http2FrameType) http2frameParser {
	if f := http2frameParsers[t]; f != nil {
		return f
	}
	return http2parseUnknownFrame
}

// A FrameHeader is the 9 byte header of all HTTP/2 frames.
//
// See http://http2.github.io/http2-spec/#FrameHeader
type http2FrameHeader struct {
	valid bool // caller can access []byte fields in the Frame

	// Type is the 1 byte frame type. There are ten standard frame
	// types, but extension frame types may be written by WriteRawFrame
	// and will be returned by ReadFrame (as UnknownFrame).
	Type http2FrameType

	// Flags are the 1 byte of 8 potential bit flags per frame.
	// They are specific to the frame type.
	Flags http2Flags

	// Length is the length of the frame, not including the 9 byte header.
	// The maximum size is one byte less than 16MB (uint24), but only
	// frames up to 16KB are allowed without peer agreement.
	Length uint32

	// StreamID is which stream this frame is for. Certain frames
	// are not stream-specific, in which case this field is 0.
	StreamID uint32
}

// Header returns h. It exists so FrameHeaders can be embedded in other
// specific frame types and implement the Frame interface.
func (h http2FrameHeader) Header() http2FrameHeader { return h }

func (h http2FrameHeader) String() string {
	var buf bytes.Buffer
	buf.WriteString("[FrameHeader ")
	h.writeDebug(&buf)
	buf.WriteByte(']')
	return buf.String()
}

func (h http2FrameHeader) writeDebug(buf *bytes.Buffer) {
	buf.WriteString(h.Type.String())
	if h.Flags != 0 {
		buf.WriteString(" flags=")
		set := 0
		for i := uint8(0); i < 8; i++ {
			if h.Flags&(1<<i) == 0 {
				continue
			}
			set++
			if set > 1 {
				buf.WriteByte('|')
			}
			name := http2flagName[h.Type][http2Flags(1<<i)]
			if name != "" {
				buf.WriteString(name)
			} else {
				fmt.Fprintf(buf, "0x%x", 1<<i)
			}
		}
	}
	if h.StreamID != 0 {
		fmt.Fprintf(buf, " stream=%d", h.StreamID)
	}
	fmt.Fprintf(buf, " len=%d", h.Length)
}

func (h *http2FrameHeader) checkValid() {
	if !h.valid {
		panic("Frame accessor called on non-owned Frame")
	}
}

func (h *http2FrameHeader) invalidate() { h.valid = false }

// frame header bytes.
// Used only by ReadFrameHeader.
var http2fhBytes = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, http2frameHeaderLen)
		return &buf
	},
}

// ReadFrameHeader reads 9 bytes from r and returns a FrameHeader.
// Most users should use Framer.ReadFrame instead.
func http2ReadFrameHeader(r io.Reader) (http2FrameHeader, error) {
	bufp := http2fhBytes.Get().(*[]byte)
	defer http2fhBytes.Put(bufp)
	return http2readFrameHeader(*bufp, r)
}

func http2readFrameHeader(buf []byte, r io.Reader) (http2FrameHeader, error) {
	_, err := io.ReadFull(r, buf[:http2frameHeaderLen])
	if err != nil {
		return http2FrameHeader{}, err
	}
	return http2FrameHeader{
		Length:   (uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])),
		Type:     http2FrameType(buf[3]),
		Flags:    http2Flags(buf[4]),
		StreamID: binary.BigEndian.Uint32(buf[5:]) & (1<<31 - 1),
		valid:    true,
	}, nil
}

// A Frame is the base interface implemented by all frame types.
// Callers will generally type-assert the specific frame type:
// *HeadersFrame, *SettingsFrame, *WindowUpdateFrame, etc.
//
// Frames are only valid until the next call to Framer.ReadFrame.
type http2Frame interface {
	Header() http2FrameHeader

	// invalidate is called by Framer.ReadFrame to make this
	// frame's buffers as being invalid, since the subsequent
	// frame will reuse them.
	invalidate()
}

// A Framer reads and writes Frames.
type http2Framer struct {
	cc        *http2ClientConn
	r         io.Reader
	lastFrame http2Frame
	errDetail error

	// countError is a non-nil func that's called on a frame parse
	// error with some unique error path token. It's initialized
	// from Transport.CountError or Server.CountError.
	countError func(errToken string)

	// lastHeaderStream is non-zero if the last frame was an
	// unfinished HEADERS/CONTINUATION.
	lastHeaderStream uint32

	maxReadSize uint32
	headerBuf   [http2frameHeaderLen]byte

	// TODO: let getReadBuf be configurable, and use a less memory-pinning
	// allocator in server.go to minimize memory pinned for many idle conns.
	// Will probably also need to make frame invalidation have a hook too.
	getReadBuf func(size uint32) []byte
	readBuf    []byte // cache for default getReadBuf

	maxWriteSize uint32 // zero means unlimited; TODO: implement

	w    io.Writer
	wbuf []byte

	// AllowIllegalWrites permits the Framer's Write methods to
	// write frames that do not conform to the HTTP/2 spec. This
	// permits using the Framer to test other HTTP/2
	// implementations' conformance to the spec.
	// If false, the Write methods will prefer to return an error
	// rather than comply.
	AllowIllegalWrites bool

	// AllowIllegalReads permits the Framer's ReadFrame method
	// to return non-compliant frames or frame orders.
	// This is for testing and permits using the Framer to test
	// other HTTP/2 implementations' conformance to the spec.
	// It is not compatible with ReadMetaHeaders.
	AllowIllegalReads bool

	// ReadMetaHeaders if non-nil causes ReadFrame to merge
	// HEADERS and CONTINUATION frames together and return
	// MetaHeadersFrame instead.
	ReadMetaHeaders *hpack.Decoder

	// MaxHeaderListSize is the http2 MAX_HEADER_LIST_SIZE.
	// It's used only if ReadMetaHeaders is set; 0 means a sane default
	// (currently 16MB)
	// If the limit is hit, MetaHeadersFrame.Truncated is set true.
	MaxHeaderListSize uint32

	// TODO: track which type of frame & with which flags was sent
	// last. Then return an error (unless AllowIllegalWrites) if
	// we're in the middle of a header block and a
	// non-Continuation or Continuation on a different stream is
	// attempted to be written.

	logReads, logWrites bool

	debugFramer       *http2Framer // only use for logging written writes
	debugFramerBuf    *bytes.Buffer
	debugReadLoggerf  func(string, ...interface{})
	debugWriteLoggerf func(string, ...interface{})

	frameCache *http2frameCache // nil if frames aren't reused (default)
}

func (fr *http2Framer) maxHeaderListSize() uint32 {
	if fr.MaxHeaderListSize == 0 {
		return 16 << 20 // sane default, per docs
	}
	return fr.MaxHeaderListSize
}

func (f *http2Framer) startWrite(ftype http2FrameType, flags http2Flags, streamID uint32) {
	// Write the FrameHeader.
	f.wbuf = append(f.wbuf[:0],
		0, // 3 bytes of length, filled in in endWrite
		0,
		0,
		byte(ftype),
		byte(flags),
		byte(streamID>>24),
		byte(streamID>>16),
		byte(streamID>>8),
		byte(streamID))
}

func (f *http2Framer) endWrite() error {
	// Now that we know the final size, fill in the FrameHeader in
	// the space previously reserved for it. Abuse append.
	length := len(f.wbuf) - http2frameHeaderLen
	if length >= (1 << 24) {
		return http2ErrFrameTooLarge
	}
	_ = append(f.wbuf[:0],
		byte(length>>16),
		byte(length>>8),
		byte(length))
	if f.logWrites {
		f.logWrite()
	}

	n, err := f.w.Write(f.wbuf)
	if err == nil && n != len(f.wbuf) {
		err = io.ErrShortWrite
	}
	return err
}

func (f *http2Framer) logWrite() {
	if f.debugFramer == nil {
		f.debugFramerBuf = new(bytes.Buffer)
		f.debugFramer = http2NewFramer(nil, f.debugFramerBuf)
		f.debugFramer.logReads = false // we log it ourselves, saying "wrote" below
		// Let us read anything, even if we accidentally wrote it
		// in the wrong order:
		f.debugFramer.AllowIllegalReads = true
	}
	f.debugFramerBuf.Write(f.wbuf)
	fr, err := f.debugFramer.ReadFrame()
	if err != nil {
		f.debugWriteLoggerf("http2: Framer %p: failed to decode just-written frame", f)
		return
	}
	f.debugWriteLoggerf("http2: Framer %p: wrote %v", f, http2summarizeFrame(fr))
}

func (f *http2Framer) writeByte(v byte) { f.wbuf = append(f.wbuf, v) }

func (f *http2Framer) writeBytes(v []byte) { f.wbuf = append(f.wbuf, v...) }

func (f *http2Framer) writeUint16(v uint16) { f.wbuf = append(f.wbuf, byte(v>>8), byte(v)) }

func (f *http2Framer) writeUint32(v uint32) {
	f.wbuf = append(f.wbuf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

const (
	http2minMaxFrameSize = 1 << 14
	http2maxFrameSize    = 1<<24 - 1
)

// SetReuseFrames allows the Framer to reuse Frames.
// If called on a Framer, Frames returned by calls to ReadFrame are only
// valid until the next call to ReadFrame.
func (fr *http2Framer) SetReuseFrames() {
	if fr.frameCache != nil {
		return
	}
	fr.frameCache = &http2frameCache{}
}

type http2frameCache struct {
	dataFrame http2DataFrame
}

func (fc *http2frameCache) getDataFrame() *http2DataFrame {
	if fc == nil {
		return &http2DataFrame{}
	}
	return &fc.dataFrame
}

// NewFramer returns a Framer that writes frames to w and reads them from r.
func http2NewFramer(w io.Writer, r io.Reader) *http2Framer {
	fr := &http2Framer{
		w:                 w,
		r:                 r,
		countError:        func(string) {},
		logReads:          http2logFrameReads,
		logWrites:         http2logFrameWrites,
		debugReadLoggerf:  log.Printf,
		debugWriteLoggerf: log.Printf,
	}
	fr.getReadBuf = func(size uint32) []byte {
		if cap(fr.readBuf) >= int(size) {
			return fr.readBuf[:size]
		}
		fr.readBuf = make([]byte, size)
		return fr.readBuf
	}
	fr.SetMaxReadFrameSize(http2maxFrameSize)
	return fr
}

// SetMaxReadFrameSize sets the maximum size of a frame
// that will be read by a subsequent call to ReadFrame.
// It is the caller's responsibility to advertise this
// limit with a SETTINGS frame.
func (fr *http2Framer) SetMaxReadFrameSize(v uint32) {
	if v > http2maxFrameSize {
		v = http2maxFrameSize
	}
	fr.maxReadSize = v
}

// ErrorDetail returns a more detailed error of the last error
// returned by Framer.ReadFrame. For instance, if ReadFrame
// returns a StreamError with code PROTOCOL_ERROR, ErrorDetail
// will say exactly what was invalid. ErrorDetail is not guaranteed
// to return a non-nil value and like the rest of the http2 package,
// its return value is not protected by an API compatibility promise.
// ErrorDetail is reset after the next call to ReadFrame.
func (fr *http2Framer) ErrorDetail() error {
	return fr.errDetail
}

// ErrFrameTooLarge is returned from Framer.ReadFrame when the peer
// sends a frame that is larger than declared with SetMaxReadFrameSize.
var http2ErrFrameTooLarge = errors.New("http2: frame too large")

// terminalReadFrameError reports whether err is an unrecoverable
// error from ReadFrame and no other frames should be read.
func http2terminalReadFrameError(err error) bool {
	if _, ok := err.(http2StreamError); ok {
		return false
	}
	return err != nil
}

// ReadFrame reads a single frame. The returned Frame is only valid
// until the next call to ReadFrame.
//
// If the frame is larger than previously set with SetMaxReadFrameSize, the
// returned error is ErrFrameTooLarge. Other errors may be of type
// ConnectionError, StreamError, or anything else from the underlying
// reader.
func (fr *http2Framer) ReadFrame() (http2Frame, error) {
	fr.errDetail = nil
	if fr.lastFrame != nil {
		fr.lastFrame.invalidate()
	}
	fh, err := http2readFrameHeader(fr.headerBuf[:], fr.r)
	if err != nil {
		return nil, err
	}
	if fh.Length > fr.maxReadSize {
		return nil, http2ErrFrameTooLarge
	}
	payload := fr.getReadBuf(fh.Length)
	if _, err := io.ReadFull(fr.r, payload); err != nil {
		return nil, err
	}
	f, err := http2typeFrameParser(fh.Type)(fr.frameCache, fh, fr.countError, payload)
	if err != nil {
		if ce, ok := err.(http2connError); ok {
			return nil, fr.connError(ce.Code, ce.Reason)
		}
		return nil, err
	}
	if err := fr.checkFrameOrder(f); err != nil {
		return nil, err
	}
	if fr.logReads {
		fr.debugReadLoggerf("http2: Framer %p: read %v", fr, http2summarizeFrame(f))
	}
	if fh.Type == http2FrameHeaders && fr.ReadMetaHeaders != nil {
		var dumps []*dumper
		if fr.cc != nil {
			dumps = getDumpers(fr.cc.t.t1.dump, fr.cc.currentRequest.Context())
		}
		if len(dumps) > 0 {
			dd := []*dumper{}
			for _, dump := range dumps {
				if dump.ResponseHeader {
					dd = append(dd, dump)
				}
			}
			dumps = dd
		}
		hr, err := fr.readMetaFrame(f.(*http2HeadersFrame), dumps)
		if err == nil && len(dumps) > 0 {
			for _, dump := range dumps {
				dump.dump([]byte("\r\n"))
			}
		}
		return hr, err
	}
	return f, nil
}

// connError returns ConnectionError(code) but first
// stashes away a public reason to the caller can optionally relay it
// to the peer before hanging up on them. This might help others debug
// their implementations.
func (fr *http2Framer) connError(code http2ErrCode, reason string) error {
	fr.errDetail = errors.New(reason)
	return http2ConnectionError(code)
}

// checkFrameOrder reports an error if f is an invalid frame to return
// next from ReadFrame. Mostly it checks whether HEADERS and
// CONTINUATION frames are contiguous.
func (fr *http2Framer) checkFrameOrder(f http2Frame) error {
	last := fr.lastFrame
	fr.lastFrame = f
	if fr.AllowIllegalReads {
		return nil
	}

	fh := f.Header()
	if fr.lastHeaderStream != 0 {
		if fh.Type != http2FrameContinuation {
			return fr.connError(http2ErrCodeProtocol,
				fmt.Sprintf("got %s for stream %d; expected CONTINUATION following %s for stream %d",
					fh.Type, fh.StreamID,
					last.Header().Type, fr.lastHeaderStream))
		}
		if fh.StreamID != fr.lastHeaderStream {
			return fr.connError(http2ErrCodeProtocol,
				fmt.Sprintf("got CONTINUATION for stream %d; expected stream %d",
					fh.StreamID, fr.lastHeaderStream))
		}
	} else if fh.Type == http2FrameContinuation {
		return fr.connError(http2ErrCodeProtocol, fmt.Sprintf("unexpected CONTINUATION for stream %d", fh.StreamID))
	}

	switch fh.Type {
	case http2FrameHeaders, http2FrameContinuation:
		if fh.Flags.Has(http2FlagHeadersEndHeaders) {
			fr.lastHeaderStream = 0
		} else {
			fr.lastHeaderStream = fh.StreamID
		}
	}

	return nil
}

// A DataFrame conveys arbitrary, variable-length sequences of octets
// associated with a stream.
// See http://http2.github.io/http2-spec/#rfc.section.6.1
type http2DataFrame struct {
	http2FrameHeader
	data []byte
}

func (f *http2DataFrame) StreamEnded() bool {
	return f.http2FrameHeader.Flags.Has(http2FlagDataEndStream)
}

// Data returns the frame's data octets, not including any padding
// size byte or padding suffix bytes.
// The caller must not retain the returned memory past the next
// call to ReadFrame.
func (f *http2DataFrame) Data() []byte {
	f.checkValid()
	return f.data
}

func http2parseDataFrame(fc *http2frameCache, fh http2FrameHeader, countError func(string), payload []byte) (http2Frame, error) {
	if fh.StreamID == 0 {
		// DATA frames MUST be associated with a stream. If a
		// DATA frame is received whose stream identifier
		// field is 0x0, the recipient MUST respond with a
		// connection error (Section 5.4.1) of type
		// PROTOCOL_ERROR.
		countError("frame_data_stream_0")
		return nil, http2connError{http2ErrCodeProtocol, "DATA frame with stream ID 0"}
	}
	f := fc.getDataFrame()
	f.http2FrameHeader = fh

	var padSize byte
	if fh.Flags.Has(http2FlagDataPadded) {
		var err error
		payload, padSize, err = http2readByte(payload)
		if err != nil {
			countError("frame_data_pad_byte_short")
			return nil, err
		}
	}
	if int(padSize) > len(payload) {
		// If the length of the padding is greater than the
		// length of the frame payload, the recipient MUST
		// treat this as a connection error.
		// Filed: https://github.com/http2/http2-spec/issues/610
		countError("frame_data_pad_too_big")
		return nil, http2connError{http2ErrCodeProtocol, "pad size larger than data payload"}
	}
	f.data = payload[:len(payload)-int(padSize)]
	return f, nil
}

var (
	http2errStreamID    = errors.New("invalid stream ID")
	http2errDepStreamID = errors.New("invalid dependent stream ID")
	http2errPadLength   = errors.New("pad length too large")
	http2errPadBytes    = errors.New("padding bytes must all be zeros unless AllowIllegalWrites is enabled")
)

func http2validStreamIDOrZero(streamID uint32) bool {
	return streamID&(1<<31) == 0
}

func http2validStreamID(streamID uint32) bool {
	return streamID != 0 && streamID&(1<<31) == 0
}

// writeData writes a DATA frame.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility not to violate the maximum frame size
// and to not call other Write methods concurrently.
func (f *http2Framer) WriteData(streamID uint32, endStream bool, data []byte) error {
	return f.WriteDataPadded(streamID, endStream, data, nil)
}

// WriteDataPadded writes a DATA frame with optional padding.
//
// If pad is nil, the padding bit is not sent.
// The length of pad must not exceed 255 bytes.
// The bytes of pad must all be zero, unless f.AllowIllegalWrites is set.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility not to violate the maximum frame size
// and to not call other Write methods concurrently.
func (f *http2Framer) WriteDataPadded(streamID uint32, endStream bool, data, pad []byte) error {
	if !http2validStreamID(streamID) && !f.AllowIllegalWrites {
		return http2errStreamID
	}
	if len(pad) > 0 {
		if len(pad) > 255 {
			return http2errPadLength
		}
		if !f.AllowIllegalWrites {
			for _, b := range pad {
				if b != 0 {
					// "Padding octets MUST be set to zero when sending."
					return http2errPadBytes
				}
			}
		}
	}
	var flags http2Flags
	if endStream {
		flags |= http2FlagDataEndStream
	}
	if pad != nil {
		flags |= http2FlagDataPadded
	}
	f.startWrite(http2FrameData, flags, streamID)
	if pad != nil {
		f.wbuf = append(f.wbuf, byte(len(pad)))
	}
	f.wbuf = append(f.wbuf, data...)
	f.wbuf = append(f.wbuf, pad...)
	return f.endWrite()
}

// A SettingsFrame conveys configuration parameters that affect how
// endpoints communicate, such as preferences and constraints on peer
// behavior.
//
// See http://http2.github.io/http2-spec/#SETTINGS
type http2SettingsFrame struct {
	http2FrameHeader
	p []byte
}

func http2parseSettingsFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (http2Frame, error) {
	if fh.Flags.Has(http2FlagSettingsAck) && fh.Length > 0 {
		// When this (ACK 0x1) bit is set, the payload of the
		// SETTINGS frame MUST be empty. Receipt of a
		// SETTINGS frame with the ACK flag set and a length
		// field value other than 0 MUST be treated as a
		// connection error (Section 5.4.1) of type
		// FRAME_SIZE_ERROR.
		countError("frame_settings_ack_with_length")
		return nil, http2ConnectionError(http2ErrCodeFrameSize)
	}
	if fh.StreamID != 0 {
		// SETTINGS frames always apply to a connection,
		// never a single stream. The stream identifier for a
		// SETTINGS frame MUST be zero (0x0).  If an endpoint
		// receives a SETTINGS frame whose stream identifier
		// field is anything other than 0x0, the endpoint MUST
		// respond with a connection error (Section 5.4.1) of
		// type PROTOCOL_ERROR.
		countError("frame_settings_has_stream")
		return nil, http2ConnectionError(http2ErrCodeProtocol)
	}
	if len(p)%6 != 0 {
		countError("frame_settings_mod_6")
		// Expecting even number of 6 byte settings.
		return nil, http2ConnectionError(http2ErrCodeFrameSize)
	}
	f := &http2SettingsFrame{http2FrameHeader: fh, p: p}
	if v, ok := f.Value(http2SettingInitialWindowSize); ok && v > (1<<31)-1 {
		countError("frame_settings_window_size_too_big")
		// Values above the maximum flow control window size of 2^31 - 1 MUST
		// be treated as a connection error (Section 5.4.1) of type
		// FLOW_CONTROL_ERROR.
		return nil, http2ConnectionError(http2ErrCodeFlowControl)
	}
	return f, nil
}

func (f *http2SettingsFrame) IsAck() bool {
	return f.http2FrameHeader.Flags.Has(http2FlagSettingsAck)
}

func (f *http2SettingsFrame) Value(id http2SettingID) (v uint32, ok bool) {
	f.checkValid()
	for i := 0; i < f.NumSettings(); i++ {
		if s := f.Setting(i); s.ID == id {
			return s.Val, true
		}
	}
	return 0, false
}

// Setting returns the setting from the frame at the given 0-based index.
// The index must be >= 0 and less than f.NumSettings().
func (f *http2SettingsFrame) Setting(i int) http2Setting {
	buf := f.p
	return http2Setting{
		ID:  http2SettingID(binary.BigEndian.Uint16(buf[i*6 : i*6+2])),
		Val: binary.BigEndian.Uint32(buf[i*6+2 : i*6+6]),
	}
}

func (f *http2SettingsFrame) NumSettings() int { return len(f.p) / 6 }

// HasDuplicates reports whether f contains any duplicate setting IDs.
func (f *http2SettingsFrame) HasDuplicates() bool {
	num := f.NumSettings()
	if num == 0 {
		return false
	}
	// If it's small enough (the common case), just do the n^2
	// thing and avoid a map allocation.
	if num < 10 {
		for i := 0; i < num; i++ {
			idi := f.Setting(i).ID
			for j := i + 1; j < num; j++ {
				idj := f.Setting(j).ID
				if idi == idj {
					return true
				}
			}
		}
		return false
	}
	seen := map[http2SettingID]bool{}
	for i := 0; i < num; i++ {
		id := f.Setting(i).ID
		if seen[id] {
			return true
		}
		seen[id] = true
	}
	return false
}

// ForeachSetting runs fn for each setting.
// It stops and returns the first error.
func (f *http2SettingsFrame) ForeachSetting(fn func(http2Setting) error) error {
	f.checkValid()
	for i := 0; i < f.NumSettings(); i++ {
		if err := fn(f.Setting(i)); err != nil {
			return err
		}
	}
	return nil
}

// WriteSettings writes a SETTINGS frame with zero or more settings
// specified and the ACK bit not set.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *http2Framer) WriteSettings(settings ...http2Setting) error {
	f.startWrite(http2FrameSettings, 0, 0)
	for _, s := range settings {
		f.writeUint16(uint16(s.ID))
		f.writeUint32(s.Val)
	}
	return f.endWrite()
}

// WriteSettingsAck writes an empty SETTINGS frame with the ACK bit set.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *http2Framer) WriteSettingsAck() error {
	f.startWrite(http2FrameSettings, http2FlagSettingsAck, 0)
	return f.endWrite()
}

// A PingFrame is a mechanism for measuring a minimal round trip time
// from the sender, as well as determining whether an idle connection
// is still functional.
// See http://http2.github.io/http2-spec/#rfc.section.6.7
type http2PingFrame struct {
	http2FrameHeader
	Data [8]byte
}

func (f *http2PingFrame) IsAck() bool { return f.Flags.Has(http2FlagPingAck) }

func http2parsePingFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), payload []byte) (http2Frame, error) {
	if len(payload) != 8 {
		countError("frame_ping_length")
		return nil, http2ConnectionError(http2ErrCodeFrameSize)
	}
	if fh.StreamID != 0 {
		countError("frame_ping_has_stream")
		return nil, http2ConnectionError(http2ErrCodeProtocol)
	}
	f := &http2PingFrame{http2FrameHeader: fh}
	copy(f.Data[:], payload)
	return f, nil
}

func (f *http2Framer) WritePing(ack bool, data [8]byte) error {
	var flags http2Flags
	if ack {
		flags = http2FlagPingAck
	}
	f.startWrite(http2FramePing, flags, 0)
	f.writeBytes(data[:])
	return f.endWrite()
}

// A GoAwayFrame informs the remote peer to stop creating streams on this connection.
// See http://http2.github.io/http2-spec/#rfc.section.6.8
type http2GoAwayFrame struct {
	http2FrameHeader
	LastStreamID uint32
	ErrCode      http2ErrCode
	debugData    []byte
}

// DebugData returns any debug data in the GOAWAY frame. Its contents
// are not defined.
// The caller must not retain the returned memory past the next
// call to ReadFrame.
func (f *http2GoAwayFrame) DebugData() []byte {
	f.checkValid()
	return f.debugData
}

func http2parseGoAwayFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (http2Frame, error) {
	if fh.StreamID != 0 {
		countError("frame_goaway_has_stream")
		return nil, http2ConnectionError(http2ErrCodeProtocol)
	}
	if len(p) < 8 {
		countError("frame_goaway_short")
		return nil, http2ConnectionError(http2ErrCodeFrameSize)
	}
	return &http2GoAwayFrame{
		http2FrameHeader: fh,
		LastStreamID:     binary.BigEndian.Uint32(p[:4]) & (1<<31 - 1),
		ErrCode:          http2ErrCode(binary.BigEndian.Uint32(p[4:8])),
		debugData:        p[8:],
	}, nil
}

func (f *http2Framer) WriteGoAway(maxStreamID uint32, code http2ErrCode, debugData []byte) error {
	f.startWrite(http2FrameGoAway, 0, 0)
	f.writeUint32(maxStreamID & (1<<31 - 1))
	f.writeUint32(uint32(code))
	f.writeBytes(debugData)
	return f.endWrite()
}

// An UnknownFrame is the frame type returned when the frame type is unknown
// or no specific frame type parser exists.
type http2UnknownFrame struct {
	http2FrameHeader
	p []byte
}

// Payload returns the frame's payload (after the header).  It is not
// valid to call this method after a subsequent call to
// Framer.ReadFrame, nor is it valid to retain the returned slice.
// The memory is owned by the Framer and is invalidated when the next
// frame is read.
func (f *http2UnknownFrame) Payload() []byte {
	f.checkValid()
	return f.p
}

func http2parseUnknownFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (http2Frame, error) {
	return &http2UnknownFrame{fh, p}, nil
}

// A WindowUpdateFrame is used to implement flow control.
// See http://http2.github.io/http2-spec/#rfc.section.6.9
type http2WindowUpdateFrame struct {
	http2FrameHeader
	Increment uint32 // never read with high bit set
}

func http2parseWindowUpdateFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (http2Frame, error) {
	if len(p) != 4 {
		countError("frame_windowupdate_bad_len")
		return nil, http2ConnectionError(http2ErrCodeFrameSize)
	}
	inc := binary.BigEndian.Uint32(p[:4]) & 0x7fffffff // mask off high reserved bit
	if inc == 0 {
		// A receiver MUST treat the receipt of a
		// WINDOW_UPDATE frame with an flow control window
		// increment of 0 as a stream error (Section 5.4.2) of
		// type PROTOCOL_ERROR; errors on the connection flow
		// control window MUST be treated as a connection
		// error (Section 5.4.1).
		if fh.StreamID == 0 {
			countError("frame_windowupdate_zero_inc_conn")
			return nil, http2ConnectionError(http2ErrCodeProtocol)
		}
		countError("frame_windowupdate_zero_inc_stream")
		return nil, http2streamError(fh.StreamID, http2ErrCodeProtocol)
	}
	return &http2WindowUpdateFrame{
		http2FrameHeader: fh,
		Increment:        inc,
	}, nil
}

// WriteWindowUpdate writes a WINDOW_UPDATE frame.
// The increment value must be between 1 and 2,147,483,647, inclusive.
// If the Stream ID is zero, the window update applies to the
// connection as a whole.
func (f *http2Framer) WriteWindowUpdate(streamID, incr uint32) error {
	// "The legal range for the increment to the flow control window is 1 to 2^31-1 (2,147,483,647) octets."
	if (incr < 1 || incr > 2147483647) && !f.AllowIllegalWrites {
		return errors.New("illegal window increment value")
	}
	f.startWrite(http2FrameWindowUpdate, 0, streamID)
	f.writeUint32(incr)
	return f.endWrite()
}

// A HeadersFrame is used to open a stream and additionally carries a
// header block fragment.
type http2HeadersFrame struct {
	http2FrameHeader

	// Priority is set if FlagHeadersPriority is set in the FrameHeader.
	Priority http2PriorityParam

	headerFragBuf []byte // not owned
}

func (f *http2HeadersFrame) HeaderBlockFragment() []byte {
	f.checkValid()
	return f.headerFragBuf
}

func (f *http2HeadersFrame) HeadersEnded() bool {
	return f.http2FrameHeader.Flags.Has(http2FlagHeadersEndHeaders)
}

func (f *http2HeadersFrame) StreamEnded() bool {
	return f.http2FrameHeader.Flags.Has(http2FlagHeadersEndStream)
}

func (f *http2HeadersFrame) HasPriority() bool {
	return f.http2FrameHeader.Flags.Has(http2FlagHeadersPriority)
}

func http2parseHeadersFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (_ http2Frame, err error) {
	hf := &http2HeadersFrame{
		http2FrameHeader: fh,
	}
	if fh.StreamID == 0 {
		// HEADERS frames MUST be associated with a stream. If a HEADERS frame
		// is received whose stream identifier field is 0x0, the recipient MUST
		// respond with a connection error (Section 5.4.1) of type
		// PROTOCOL_ERROR.
		countError("frame_headers_zero_stream")
		return nil, http2connError{http2ErrCodeProtocol, "HEADERS frame with stream ID 0"}
	}
	var padLength uint8
	if fh.Flags.Has(http2FlagHeadersPadded) {
		if p, padLength, err = http2readByte(p); err != nil {
			countError("frame_headers_pad_short")
			return
		}
	}
	if fh.Flags.Has(http2FlagHeadersPriority) {
		var v uint32
		p, v, err = http2readUint32(p)
		if err != nil {
			countError("frame_headers_prio_short")
			return nil, err
		}
		hf.Priority.StreamDep = v & 0x7fffffff
		hf.Priority.Exclusive = (v != hf.Priority.StreamDep) // high bit was set
		p, hf.Priority.Weight, err = http2readByte(p)
		if err != nil {
			countError("frame_headers_prio_weight_short")
			return nil, err
		}
	}
	if len(p)-int(padLength) < 0 {
		countError("frame_headers_pad_too_big")
		return nil, http2streamError(fh.StreamID, http2ErrCodeProtocol)
	}
	hf.headerFragBuf = p[:len(p)-int(padLength)]
	return hf, nil
}

// HeadersFrameParam are the parameters for writing a HEADERS frame.
type http2HeadersFrameParam struct {
	// StreamID is the required Stream ID to initiate.
	StreamID uint32
	// BlockFragment is part (or all) of a Header Block.
	BlockFragment []byte

	// EndStream indicates that the header block is the last that
	// the endpoint will send for the identified stream. Setting
	// this flag causes the stream to enter one of "half closed"
	// states.
	EndStream bool

	// EndHeaders indicates that this frame contains an entire
	// header block and is not followed by any
	// CONTINUATION frames.
	EndHeaders bool

	// PadLength is the optional number of bytes of zeros to add
	// to this frame.
	PadLength uint8

	// Priority, if non-zero, includes stream priority information
	// in the HEADER frame.
	Priority http2PriorityParam
}

// WriteHeaders writes a single HEADERS frame.
//
// This is a low-level header writing method. Encoding headers and
// splitting them into any necessary CONTINUATION frames is handled
// elsewhere.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *http2Framer) WriteHeaders(p http2HeadersFrameParam) error {
	if !http2validStreamID(p.StreamID) && !f.AllowIllegalWrites {
		return http2errStreamID
	}
	var flags http2Flags
	if p.PadLength != 0 {
		flags |= http2FlagHeadersPadded
	}
	if p.EndStream {
		flags |= http2FlagHeadersEndStream
	}
	if p.EndHeaders {
		flags |= http2FlagHeadersEndHeaders
	}
	if !p.Priority.IsZero() {
		flags |= http2FlagHeadersPriority
	}
	f.startWrite(http2FrameHeaders, flags, p.StreamID)
	if p.PadLength != 0 {
		f.writeByte(p.PadLength)
	}
	if !p.Priority.IsZero() {
		v := p.Priority.StreamDep
		if !http2validStreamIDOrZero(v) && !f.AllowIllegalWrites {
			return http2errDepStreamID
		}
		if p.Priority.Exclusive {
			v |= 1 << 31
		}
		f.writeUint32(v)
		f.writeByte(p.Priority.Weight)
	}
	f.wbuf = append(f.wbuf, p.BlockFragment...)
	f.wbuf = append(f.wbuf, http2padZeros[:p.PadLength]...)
	return f.endWrite()
}

// A PriorityFrame specifies the sender-advised priority of a stream.
// See http://http2.github.io/http2-spec/#rfc.section.6.3
type http2PriorityFrame struct {
	http2FrameHeader
	http2PriorityParam
}

// PriorityParam are the stream prioritzation parameters.
type http2PriorityParam struct {
	// StreamDep is a 31-bit stream identifier for the
	// stream that this stream depends on. Zero means no
	// dependency.
	StreamDep uint32

	// Exclusive is whether the dependency is exclusive.
	Exclusive bool

	// Weight is the stream's zero-indexed weight. It should be
	// set together with StreamDep, or neither should be set. Per
	// the spec, "Add one to the value to obtain a weight between
	// 1 and 256."
	Weight uint8
}

func (p http2PriorityParam) IsZero() bool {
	return p == http2PriorityParam{}
}

func http2parsePriorityFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), payload []byte) (http2Frame, error) {
	if fh.StreamID == 0 {
		countError("frame_priority_zero_stream")
		return nil, http2connError{http2ErrCodeProtocol, "PRIORITY frame with stream ID 0"}
	}
	if len(payload) != 5 {
		countError("frame_priority_bad_length")
		return nil, http2connError{http2ErrCodeFrameSize, fmt.Sprintf("PRIORITY frame payload size was %d; want 5", len(payload))}
	}
	v := binary.BigEndian.Uint32(payload[:4])
	streamID := v & 0x7fffffff // mask off high bit
	return &http2PriorityFrame{
		http2FrameHeader: fh,
		http2PriorityParam: http2PriorityParam{
			Weight:    payload[4],
			StreamDep: streamID,
			Exclusive: streamID != v, // was high bit set?
		},
	}, nil
}

// WritePriority writes a PRIORITY frame.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *http2Framer) WritePriority(streamID uint32, p http2PriorityParam) error {
	if !http2validStreamID(streamID) && !f.AllowIllegalWrites {
		return http2errStreamID
	}
	if !http2validStreamIDOrZero(p.StreamDep) {
		return http2errDepStreamID
	}
	f.startWrite(http2FramePriority, 0, streamID)
	v := p.StreamDep
	if p.Exclusive {
		v |= 1 << 31
	}
	f.writeUint32(v)
	f.writeByte(p.Weight)
	return f.endWrite()
}

// A RSTStreamFrame allows for abnormal termination of a stream.
// See http://http2.github.io/http2-spec/#rfc.section.6.4
type http2RSTStreamFrame struct {
	http2FrameHeader
	ErrCode http2ErrCode
}

func http2parseRSTStreamFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (http2Frame, error) {
	if len(p) != 4 {
		countError("frame_rststream_bad_len")
		return nil, http2ConnectionError(http2ErrCodeFrameSize)
	}
	if fh.StreamID == 0 {
		countError("frame_rststream_zero_stream")
		return nil, http2ConnectionError(http2ErrCodeProtocol)
	}
	return &http2RSTStreamFrame{fh, http2ErrCode(binary.BigEndian.Uint32(p[:4]))}, nil
}

// WriteRSTStream writes a RST_STREAM frame.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *http2Framer) WriteRSTStream(streamID uint32, code http2ErrCode) error {
	if !http2validStreamID(streamID) && !f.AllowIllegalWrites {
		return http2errStreamID
	}
	f.startWrite(http2FrameRSTStream, 0, streamID)
	f.writeUint32(uint32(code))
	return f.endWrite()
}

// A ContinuationFrame is used to continue a sequence of header block fragments.
// See http://http2.github.io/http2-spec/#rfc.section.6.10
type http2ContinuationFrame struct {
	http2FrameHeader
	headerFragBuf []byte
}

func http2parseContinuationFrame(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (http2Frame, error) {
	if fh.StreamID == 0 {
		countError("frame_continuation_zero_stream")
		return nil, http2connError{http2ErrCodeProtocol, "CONTINUATION frame with stream ID 0"}
	}
	return &http2ContinuationFrame{fh, p}, nil
}

func (f *http2ContinuationFrame) HeaderBlockFragment() []byte {
	f.checkValid()
	return f.headerFragBuf
}

func (f *http2ContinuationFrame) HeadersEnded() bool {
	return f.http2FrameHeader.Flags.Has(http2FlagContinuationEndHeaders)
}

// WriteContinuation writes a CONTINUATION frame.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *http2Framer) WriteContinuation(streamID uint32, endHeaders bool, headerBlockFragment []byte) error {
	if !http2validStreamID(streamID) && !f.AllowIllegalWrites {
		return http2errStreamID
	}
	var flags http2Flags
	if endHeaders {
		flags |= http2FlagContinuationEndHeaders
	}
	f.startWrite(http2FrameContinuation, flags, streamID)
	f.wbuf = append(f.wbuf, headerBlockFragment...)
	return f.endWrite()
}

// A PushPromiseFrame is used to initiate a server stream.
// See http://http2.github.io/http2-spec/#rfc.section.6.6
type http2PushPromiseFrame struct {
	http2FrameHeader
	PromiseID     uint32
	headerFragBuf []byte // not owned
}

func (f *http2PushPromiseFrame) HeaderBlockFragment() []byte {
	f.checkValid()
	return f.headerFragBuf
}

func (f *http2PushPromiseFrame) HeadersEnded() bool {
	return f.http2FrameHeader.Flags.Has(http2FlagPushPromiseEndHeaders)
}

func http2parsePushPromise(_ *http2frameCache, fh http2FrameHeader, countError func(string), p []byte) (_ http2Frame, err error) {
	pp := &http2PushPromiseFrame{
		http2FrameHeader: fh,
	}
	if pp.StreamID == 0 {
		// PUSH_PROMISE frames MUST be associated with an existing,
		// peer-initiated stream. The stream identifier of a
		// PUSH_PROMISE frame indicates the stream it is associated
		// with. If the stream identifier field specifies the value
		// 0x0, a recipient MUST respond with a connection error
		// (Section 5.4.1) of type PROTOCOL_ERROR.
		countError("frame_pushpromise_zero_stream")
		return nil, http2ConnectionError(http2ErrCodeProtocol)
	}
	// The PUSH_PROMISE frame includes optional padding.
	// Padding fields and flags are identical to those defined for DATA frames
	var padLength uint8
	if fh.Flags.Has(http2FlagPushPromisePadded) {
		if p, padLength, err = http2readByte(p); err != nil {
			countError("frame_pushpromise_pad_short")
			return
		}
	}

	p, pp.PromiseID, err = http2readUint32(p)
	if err != nil {
		countError("frame_pushpromise_promiseid_short")
		return
	}
	pp.PromiseID = pp.PromiseID & (1<<31 - 1)

	if int(padLength) > len(p) {
		// like the DATA frame, error out if padding is longer than the body.
		countError("frame_pushpromise_pad_too_big")
		return nil, http2ConnectionError(http2ErrCodeProtocol)
	}
	pp.headerFragBuf = p[:len(p)-int(padLength)]
	return pp, nil
}

// PushPromiseParam are the parameters for writing a PUSH_PROMISE frame.
type http2PushPromiseParam struct {
	// StreamID is the required Stream ID to initiate.
	StreamID uint32

	// PromiseID is the required Stream ID which this
	// Push Promises
	PromiseID uint32

	// BlockFragment is part (or all) of a Header Block.
	BlockFragment []byte

	// EndHeaders indicates that this frame contains an entire
	// header block and is not followed by any
	// CONTINUATION frames.
	EndHeaders bool

	// PadLength is the optional number of bytes of zeros to add
	// to this frame.
	PadLength uint8
}

// WritePushPromise writes a single PushPromise Frame.
//
// As with Header Frames, This is the low level call for writing
// individual frames. Continuation frames are handled elsewhere.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *http2Framer) WritePushPromise(p http2PushPromiseParam) error {
	if !http2validStreamID(p.StreamID) && !f.AllowIllegalWrites {
		return http2errStreamID
	}
	var flags http2Flags
	if p.PadLength != 0 {
		flags |= http2FlagPushPromisePadded
	}
	if p.EndHeaders {
		flags |= http2FlagPushPromiseEndHeaders
	}
	f.startWrite(http2FramePushPromise, flags, p.StreamID)
	if p.PadLength != 0 {
		f.writeByte(p.PadLength)
	}
	if !http2validStreamID(p.PromiseID) && !f.AllowIllegalWrites {
		return http2errStreamID
	}
	f.writeUint32(p.PromiseID)
	f.wbuf = append(f.wbuf, p.BlockFragment...)
	f.wbuf = append(f.wbuf, http2padZeros[:p.PadLength]...)
	return f.endWrite()
}

// WriteRawFrame writes a raw frame. This can be used to write
// extension frames unknown to this package.
func (f *http2Framer) WriteRawFrame(t http2FrameType, flags http2Flags, streamID uint32, payload []byte) error {
	f.startWrite(t, flags, streamID)
	f.writeBytes(payload)
	return f.endWrite()
}

func http2readByte(p []byte) (remain []byte, b byte, err error) {
	if len(p) == 0 {
		return nil, 0, io.ErrUnexpectedEOF
	}
	return p[1:], p[0], nil
}

func http2readUint32(p []byte) (remain []byte, v uint32, err error) {
	if len(p) < 4 {
		return nil, 0, io.ErrUnexpectedEOF
	}
	return p[4:], binary.BigEndian.Uint32(p[:4]), nil
}

type http2streamEnder interface {
	StreamEnded() bool
}

type http2headersEnder interface {
	HeadersEnded() bool
}

type http2headersOrContinuation interface {
	http2headersEnder
	HeaderBlockFragment() []byte
}

// A MetaHeadersFrame is the representation of one HEADERS frame and
// zero or more contiguous CONTINUATION frames and the decoding of
// their HPACK-encoded contents.
//
// This type of frame does not appear on the wire and is only returned
// by the Framer when Framer.ReadMetaHeaders is set.
type http2MetaHeadersFrame struct {
	*http2HeadersFrame

	// Fields are the fields contained in the HEADERS and
	// CONTINUATION frames. The underlying slice is owned by the
	// Framer and must not be retained after the next call to
	// ReadFrame.
	//
	// Fields are guaranteed to be in the correct http2 order and
	// not have unknown pseudo header fields or invalid header
	// field names or values. Required pseudo header fields may be
	// missing, however. Use the MetaHeadersFrame.Pseudo accessor
	// method access pseudo headers.
	Fields []hpack.HeaderField

	// Truncated is whether the max header list size limit was hit
	// and Fields is incomplete. The hpack decoder state is still
	// valid, however.
	Truncated bool
}

// PseudoValue returns the given pseudo header field's value.
// The provided pseudo field should not contain the leading colon.
func (mh *http2MetaHeadersFrame) PseudoValue(pseudo string) string {
	for _, hf := range mh.Fields {
		if !hf.IsPseudo() {
			return ""
		}
		if hf.Name[1:] == pseudo {
			return hf.Value
		}
	}
	return ""
}

// RegularFields returns the regular (non-pseudo) header fields of mh.
// The caller does not own the returned slice.
func (mh *http2MetaHeadersFrame) RegularFields() []hpack.HeaderField {
	for i, hf := range mh.Fields {
		if !hf.IsPseudo() {
			return mh.Fields[i:]
		}
	}
	return nil
}

// PseudoFields returns the pseudo header fields of mh.
// The caller does not own the returned slice.
func (mh *http2MetaHeadersFrame) PseudoFields() []hpack.HeaderField {
	for i, hf := range mh.Fields {
		if !hf.IsPseudo() {
			return mh.Fields[:i]
		}
	}
	return mh.Fields
}

func (mh *http2MetaHeadersFrame) checkPseudos() error {
	var isRequest, isResponse bool
	pf := mh.PseudoFields()
	for i, hf := range pf {
		switch hf.Name {
		case ":method", ":path", ":scheme", ":authority":
			isRequest = true
		case ":status":
			isResponse = true
		default:
			return http2pseudoHeaderError(hf.Name)
		}
		// Check for duplicates.
		// This would be a bad algorithm, but N is 4.
		// And this doesn't allocate.
		for _, hf2 := range pf[:i] {
			if hf.Name == hf2.Name {
				return http2duplicatePseudoHeaderError(hf.Name)
			}
		}
	}
	if isRequest && isResponse {
		return http2errMixPseudoHeaderTypes
	}
	return nil
}

func (fr *http2Framer) maxHeaderStringLen() int {
	v := fr.maxHeaderListSize()
	if uint32(int(v)) == v {
		return int(v)
	}
	// They had a crazy big number for MaxHeaderBytes anyway,
	// so give them unlimited header lengths:
	return 0
}

// readMetaFrame returns 0 or more CONTINUATION frames from fr and
// merge them into the provided hf and returns a MetaHeadersFrame
// with the decoded hpack values.
func (fr *http2Framer) readMetaFrame(hf *http2HeadersFrame, dumps []*dumper) (*http2MetaHeadersFrame, error) {
	if fr.AllowIllegalReads {
		return nil, errors.New("illegal use of AllowIllegalReads with ReadMetaHeaders")
	}
	mh := &http2MetaHeadersFrame{
		http2HeadersFrame: hf,
	}
	var remainSize = fr.maxHeaderListSize()
	var sawRegular bool

	var invalid error // pseudo header field errors
	hdec := fr.ReadMetaHeaders
	hdec.SetEmitEnabled(true)
	hdec.SetMaxStringLength(fr.maxHeaderStringLen())
	rawEmitFunc := func(hf hpack.HeaderField) {
		if http2VerboseLogs && fr.logReads {
			fr.debugReadLoggerf("http2: decoded hpack field %+v", hf)
		}
		if !httpguts.ValidHeaderFieldValue(hf.Value) {
			invalid = http2headerFieldValueError(hf.Value)
		}
		isPseudo := strings.HasPrefix(hf.Name, ":")
		if isPseudo {
			if sawRegular {
				invalid = http2errPseudoAfterRegular
			}
		} else {
			sawRegular = true
			if !http2validWireHeaderFieldName(hf.Name) {
				invalid = http2headerFieldNameError(hf.Name)
			}
		}

		if invalid != nil {
			hdec.SetEmitEnabled(false)
			return
		}

		size := hf.Size()
		if size > remainSize {
			hdec.SetEmitEnabled(false)
			mh.Truncated = true
			return
		}
		remainSize -= size

		mh.Fields = append(mh.Fields, hf)
	}
	emitFunc := rawEmitFunc

	if len(dumps) > 0 {
		emitFunc = func(hf hpack.HeaderField) {
			for _, dump := range dumps {
				dump.dump([]byte(fmt.Sprintf("%s: %s\r\n", hf.Name, hf.Value)))
			}
			rawEmitFunc(hf)
		}
	}

	hdec.SetEmitFunc(emitFunc)
	// Lose reference to MetaHeadersFrame:
	defer hdec.SetEmitFunc(func(hf hpack.HeaderField) {})

	var hc http2headersOrContinuation = hf
	for {
		frag := hc.HeaderBlockFragment()
		if _, err := hdec.Write(frag); err != nil {
			return nil, http2ConnectionError(http2ErrCodeCompression)
		}

		if hc.HeadersEnded() {
			break
		}
		if f, err := fr.ReadFrame(); err != nil {
			return nil, err
		} else {
			hc = f.(*http2ContinuationFrame) // guaranteed by checkFrameOrder
		}
	}

	mh.http2HeadersFrame.headerFragBuf = nil
	mh.http2HeadersFrame.invalidate()

	if err := hdec.Close(); err != nil {
		return nil, http2ConnectionError(http2ErrCodeCompression)
	}
	if invalid != nil {
		fr.errDetail = invalid
		if http2VerboseLogs {
			log.Printf("http2: invalid header: %v", invalid)
		}
		return nil, http2StreamError{mh.StreamID, http2ErrCodeProtocol, invalid}
	}
	if err := mh.checkPseudos(); err != nil {
		fr.errDetail = err
		if http2VerboseLogs {
			log.Printf("http2: invalid pseudo headers: %v", err)
		}
		return nil, http2StreamError{mh.StreamID, http2ErrCodeProtocol, err}
	}
	return mh, nil
}

func http2summarizeFrame(f http2Frame) string {
	var buf bytes.Buffer
	f.Header().writeDebug(&buf)
	switch f := f.(type) {
	case *http2SettingsFrame:
		n := 0
		f.ForeachSetting(func(s http2Setting) error {
			n++
			if n == 1 {
				buf.WriteString(", settings:")
			}
			fmt.Fprintf(&buf, " %v=%v,", s.ID, s.Val)
			return nil
		})
		if n > 0 {
			buf.Truncate(buf.Len() - 1) // remove trailing comma
		}
	case *http2DataFrame:
		data := f.Data()
		const max = 256
		if len(data) > max {
			data = data[:max]
		}
		fmt.Fprintf(&buf, " data=%q", data)
		if len(f.Data()) > max {
			fmt.Fprintf(&buf, " (%d bytes omitted)", len(f.Data())-max)
		}
	case *http2WindowUpdateFrame:
		if f.StreamID == 0 {
			buf.WriteString(" (conn)")
		}
		fmt.Fprintf(&buf, " incr=%v", f.Increment)
	case *http2PingFrame:
		fmt.Fprintf(&buf, " ping=%q", f.Data[:])
	case *http2GoAwayFrame:
		fmt.Fprintf(&buf, " LastStreamID=%v ErrCode=%v Debug=%q",
			f.LastStreamID, f.ErrCode, f.debugData)
	case *http2RSTStreamFrame:
		fmt.Fprintf(&buf, " ErrCode=%v", f.ErrCode)
	}
	return buf.String()
}
