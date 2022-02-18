// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package req

import (
	"bytes"
	"fmt"
	"github.com/imroc/req/v3/internal/tests"
	"io"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/net/http2/hpack"
)

func testFramer() (*http2Framer, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	return http2NewFramer(buf, buf), buf
}

func TestFrameSizes(t *testing.T) {
	// Catch people rearranging the FrameHeader fields.
	if got, want := int(unsafe.Sizeof(http2FrameHeader{})), 12; got != want {
		t.Errorf("FrameHeader size = %d; want %d", got, want)
	}
}

func TestFrameTypeString(t *testing.T) {
	tests := []struct {
		ft   http2FrameType
		want string
	}{
		{http2FrameData, "DATA"},
		{http2FramePing, "PING"},
		{http2FrameGoAway, "GOAWAY"},
		{0xf, "UNKNOWN_FRAME_TYPE_15"},
	}

	for i, tt := range tests {
		got := tt.ft.String()
		if got != tt.want {
			t.Errorf("%d. String(FrameType %d) = %q; want %q", i, int(tt.ft), got, tt.want)
		}
	}
}

func TestWriteRST(t *testing.T) {
	fr, buf := testFramer()
	var streamID uint32 = 1<<24 + 2<<16 + 3<<8 + 4
	var errCode uint32 = 7<<24 + 6<<16 + 5<<8 + 4
	fr.WriteRSTStream(streamID, http2ErrCode(errCode))
	const wantEnc = "\x00\x00\x04\x03\x00\x01\x02\x03\x04\x07\x06\x05\x04"
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	want := &http2RSTStreamFrame{
		http2FrameHeader: http2FrameHeader{
			valid:    true,
			Type:     0x3,
			Flags:    0x0,
			Length:   0x4,
			StreamID: 0x1020304,
		},
		ErrCode: 0x7060504,
	}
	if !reflect.DeepEqual(f, want) {
		t.Errorf("parsed back %#v; want %#v", f, want)
	}
}

func TestWriteData(t *testing.T) {
	fr, buf := testFramer()
	var streamID uint32 = 1<<24 + 2<<16 + 3<<8 + 4
	data := []byte("ABC")
	fr.WriteData(streamID, true, data)
	const wantEnc = "\x00\x00\x03\x00\x01\x01\x02\x03\x04ABC"
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	df, ok := f.(*http2DataFrame)
	if !ok {
		t.Fatalf("got %T; want *http2DataFrame", f)
	}
	if !bytes.Equal(df.Data(), data) {
		t.Errorf("got %q; want %q", df.Data(), data)
	}
	if f.Header().Flags&1 == 0 {
		t.Errorf("didn't see END_STREAM flag")
	}
}

func TestWriteDataPadded(t *testing.T) {
	tests := [...]struct {
		streamID   uint32
		endStream  bool
		data       []byte
		pad        []byte
		wantHeader http2FrameHeader
	}{
		// Unpadded:
		0: {
			streamID:  1,
			endStream: true,
			data:      []byte("foo"),
			pad:       nil,
			wantHeader: http2FrameHeader{
				Type:     http2FrameData,
				Flags:    http2FlagDataEndStream,
				Length:   3,
				StreamID: 1,
			},
		},

		// Padded bit set, but no padding:
		1: {
			streamID:  1,
			endStream: true,
			data:      []byte("foo"),
			pad:       []byte{},
			wantHeader: http2FrameHeader{
				Type:     http2FrameData,
				Flags:    http2FlagDataEndStream | http2FlagDataPadded,
				Length:   4,
				StreamID: 1,
			},
		},

		// Padded bit set, with padding:
		2: {
			streamID:  1,
			endStream: false,
			data:      []byte("foo"),
			pad:       []byte{0, 0, 0},
			wantHeader: http2FrameHeader{
				Type:     http2FrameData,
				Flags:    http2FlagDataPadded,
				Length:   7,
				StreamID: 1,
			},
		},
	}
	for i, tt := range tests {
		fr, _ := testFramer()
		fr.WriteDataPadded(tt.streamID, tt.endStream, tt.data, tt.pad)
		f, err := fr.ReadFrame()
		if err != nil {
			t.Errorf("%d. ReadFrame: %v", i, err)
			continue
		}
		got := f.Header()
		tt.wantHeader.valid = true
		if !got.Equal(tt.wantHeader) {
			t.Errorf("%d. read %+v; want %+v", i, got, tt.wantHeader)
			continue
		}
		df := f.(*http2DataFrame)
		if !bytes.Equal(df.Data(), tt.data) {
			t.Errorf("%d. got %q; want %q", i, df.Data(), tt.data)
		}
	}
}

func (fh http2FrameHeader) Equal(b http2FrameHeader) bool {
	return fh.valid == b.valid &&
		fh.Type == b.Type &&
		fh.Flags == b.Flags &&
		fh.Length == b.Length &&
		fh.StreamID == b.StreamID
}

func TestWriteHeaders(t *testing.T) {
	tests := []struct {
		name      string
		p         http2HeadersFrameParam
		wantEnc   string
		wantFrame *http2HeadersFrame
	}{
		{
			"basic",
			http2HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				Priority:      http2PriorityParam{},
			},
			"\x00\x00\x03\x01\x00\x00\x00\x00*abc",
			&http2HeadersFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: 42,
					Type:     http2FrameHeaders,
					Length:   uint32(len("abc")),
				},
				Priority:      http2PriorityParam{},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"basic + end flags",
			http2HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				Priority:      http2PriorityParam{},
			},
			"\x00\x00\x03\x01\x05\x00\x00\x00*abc",
			&http2HeadersFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: 42,
					Type:     http2FrameHeaders,
					Flags:    http2FlagHeadersEndStream | http2FlagHeadersEndHeaders,
					Length:   uint32(len("abc")),
				},
				Priority:      http2PriorityParam{},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"with padding",
			http2HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				PadLength:     5,
				Priority:      http2PriorityParam{},
			},
			"\x00\x00\t\x01\r\x00\x00\x00*\x05abc\x00\x00\x00\x00\x00",
			&http2HeadersFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: 42,
					Type:     http2FrameHeaders,
					Flags:    http2FlagHeadersEndStream | http2FlagHeadersEndHeaders | http2FlagHeadersPadded,
					Length:   uint32(1 + len("abc") + 5), // pad length + contents + padding
				},
				Priority:      http2PriorityParam{},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"with priority",
			http2HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				PadLength:     2,
				Priority: http2PriorityParam{
					StreamDep: 15,
					Exclusive: true,
					Weight:    127,
				},
			},
			"\x00\x00\v\x01-\x00\x00\x00*\x02\x80\x00\x00\x0f\u007fabc\x00\x00",
			&http2HeadersFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: 42,
					Type:     http2FrameHeaders,
					Flags:    http2FlagHeadersEndStream | http2FlagHeadersEndHeaders | http2FlagHeadersPadded | http2FlagHeadersPriority,
					Length:   uint32(1 + 5 + len("abc") + 2), // pad length + priority + contents + padding
				},
				Priority: http2PriorityParam{
					StreamDep: 15,
					Exclusive: true,
					Weight:    127,
				},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"with priority stream dep zero", // golang.org/issue/15444
			http2HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				PadLength:     2,
				Priority: http2PriorityParam{
					StreamDep: 0,
					Exclusive: true,
					Weight:    127,
				},
			},
			"\x00\x00\v\x01-\x00\x00\x00*\x02\x80\x00\x00\x00\u007fabc\x00\x00",
			&http2HeadersFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: 42,
					Type:     http2FrameHeaders,
					Flags:    http2FlagHeadersEndStream | http2FlagHeadersEndHeaders | http2FlagHeadersPadded | http2FlagHeadersPriority,
					Length:   uint32(1 + 5 + len("abc") + 2), // pad length + priority + contents + padding
				},
				Priority: http2PriorityParam{
					StreamDep: 0,
					Exclusive: true,
					Weight:    127,
				},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"zero length",
			http2HeadersFrameParam{
				StreamID: 42,
				Priority: http2PriorityParam{},
			},
			"\x00\x00\x00\x01\x00\x00\x00\x00*",
			&http2HeadersFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: 42,
					Type:     http2FrameHeaders,
					Length:   0,
				},
				Priority: http2PriorityParam{},
			},
		},
	}
	for _, tt := range tests {
		fr, buf := testFramer()
		if err := fr.WriteHeaders(tt.p); err != nil {
			t.Errorf("test %q: %v", tt.name, err)
			continue
		}
		if buf.String() != tt.wantEnc {
			t.Errorf("test %q: encoded %q; want %q", tt.name, buf.Bytes(), tt.wantEnc)
		}
		f, err := fr.ReadFrame()
		if err != nil {
			t.Errorf("test %q: failed to read the frame back: %v", tt.name, err)
			continue
		}
		if !reflect.DeepEqual(f, tt.wantFrame) {
			t.Errorf("test %q: mismatch.\n got: %#v\nwant: %#v\n", tt.name, f, tt.wantFrame)
		}
	}
}

func TestWriteInvalidStreamDep(t *testing.T) {
	fr, _ := testFramer()
	err := fr.WriteHeaders(http2HeadersFrameParam{
		StreamID: 42,
		Priority: http2PriorityParam{
			StreamDep: 1 << 31,
		},
	})
	if err != http2errDepStreamID {
		t.Errorf("header error = %v; want %q", err, http2errDepStreamID)
	}

	err = fr.WritePriority(2, http2PriorityParam{StreamDep: 1 << 31})
	if err != http2errDepStreamID {
		t.Errorf("priority error = %v; want %q", err, http2errDepStreamID)
	}
}

func TestWriteContinuation(t *testing.T) {
	const streamID = 42
	tests := []struct {
		name string
		end  bool
		frag []byte

		wantFrame *http2ContinuationFrame
	}{
		{
			"not end",
			false,
			[]byte("abc"),
			&http2ContinuationFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: streamID,
					Type:     http2FrameContinuation,
					Length:   uint32(len("abc")),
				},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"end",
			true,
			[]byte("def"),
			&http2ContinuationFrame{
				http2FrameHeader: http2FrameHeader{
					valid:    true,
					StreamID: streamID,
					Type:     http2FrameContinuation,
					Flags:    http2FlagContinuationEndHeaders,
					Length:   uint32(len("def")),
				},
				headerFragBuf: []byte("def"),
			},
		},
	}
	for _, tt := range tests {
		fr, _ := testFramer()
		if err := fr.WriteContinuation(streamID, tt.end, tt.frag); err != nil {
			t.Errorf("test %q: %v", tt.name, err)
			continue
		}
		fr.AllowIllegalReads = true
		f, err := fr.ReadFrame()
		if err != nil {
			t.Errorf("test %q: failed to read the frame back: %v", tt.name, err)
			continue
		}
		if !reflect.DeepEqual(f, tt.wantFrame) {
			t.Errorf("test %q: mismatch.\n got: %#v\nwant: %#v\n", tt.name, f, tt.wantFrame)
		}
	}
}

func TestWritePriority(t *testing.T) {
	const streamID = 42
	tests := []struct {
		name      string
		priority  http2PriorityParam
		wantFrame *http2PriorityFrame
	}{
		{
			"not exclusive",
			http2PriorityParam{
				StreamDep: 2,
				Exclusive: false,
				Weight:    127,
			},
			&http2PriorityFrame{
				http2FrameHeader{
					valid:    true,
					StreamID: streamID,
					Type:     http2FramePriority,
					Length:   5,
				},
				http2PriorityParam{
					StreamDep: 2,
					Exclusive: false,
					Weight:    127,
				},
			},
		},

		{
			"exclusive",
			http2PriorityParam{
				StreamDep: 3,
				Exclusive: true,
				Weight:    77,
			},
			&http2PriorityFrame{
				http2FrameHeader{
					valid:    true,
					StreamID: streamID,
					Type:     http2FramePriority,
					Length:   5,
				},
				http2PriorityParam{
					StreamDep: 3,
					Exclusive: true,
					Weight:    77,
				},
			},
		},
	}
	for _, tt := range tests {
		fr, _ := testFramer()
		if err := fr.WritePriority(streamID, tt.priority); err != nil {
			t.Errorf("test %q: %v", tt.name, err)
			continue
		}
		f, err := fr.ReadFrame()
		if err != nil {
			t.Errorf("test %q: failed to read the frame back: %v", tt.name, err)
			continue
		}
		if !reflect.DeepEqual(f, tt.wantFrame) {
			t.Errorf("test %q: mismatch.\n got: %#v\nwant: %#v\n", tt.name, f, tt.wantFrame)
		}
	}
}

func TestWriteSettings(t *testing.T) {
	fr, buf := testFramer()
	settings := []http2Setting{{1, 2}, {3, 4}}
	fr.WriteSettings(settings...)
	const wantEnc = "\x00\x00\f\x04\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x02\x00\x03\x00\x00\x00\x04"
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	sf, ok := f.(*http2SettingsFrame)
	if !ok {
		t.Fatalf("Got a %T; want a SettingsFrame", f)
	}
	var got []http2Setting
	sf.ForeachSetting(func(s http2Setting) error {
		got = append(got, s)
		valBack, ok := sf.Value(s.ID)
		if !ok || valBack != s.Val {
			t.Errorf("Value(%d) = %v, %v; want %v, true", s.ID, valBack, ok, s.Val)
		}
		return nil
	})
	if !reflect.DeepEqual(settings, got) {
		t.Errorf("Read settings %+v != written settings %+v", got, settings)
	}
}

func TestWriteSettingsAck(t *testing.T) {
	fr, buf := testFramer()
	fr.WriteSettingsAck()
	const wantEnc = "\x00\x00\x00\x04\x01\x00\x00\x00\x00"
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
}

func TestWriteWindowUpdate(t *testing.T) {
	fr, buf := testFramer()
	const streamID = 1<<24 + 2<<16 + 3<<8 + 4
	const incr = 7<<24 + 6<<16 + 5<<8 + 4
	if err := fr.WriteWindowUpdate(streamID, incr); err != nil {
		t.Fatal(err)
	}
	const wantEnc = "\x00\x00\x04\x08\x00\x01\x02\x03\x04\x07\x06\x05\x04"
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	want := &http2WindowUpdateFrame{
		http2FrameHeader: http2FrameHeader{
			valid:    true,
			Type:     0x8,
			Flags:    0x0,
			Length:   0x4,
			StreamID: 0x1020304,
		},
		Increment: 0x7060504,
	}
	if !reflect.DeepEqual(f, want) {
		t.Errorf("parsed back %#v; want %#v", f, want)
	}
}

func TestWritePing(t *testing.T)    { testWritePing(t, false) }
func TestWritePingAck(t *testing.T) { testWritePing(t, true) }

func testWritePing(t *testing.T, ack bool) {
	fr, buf := testFramer()
	if err := fr.WritePing(ack, [8]byte{1, 2, 3, 4, 5, 6, 7, 8}); err != nil {
		t.Fatal(err)
	}
	var wantFlags http2Flags
	if ack {
		wantFlags = http2FlagPingAck
	}
	var wantEnc = "\x00\x00\x08\x06" + string(wantFlags) + "\x00\x00\x00\x00" + "\x01\x02\x03\x04\x05\x06\x07\x08"
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}

	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	want := &http2PingFrame{
		http2FrameHeader: http2FrameHeader{
			valid:    true,
			Type:     0x6,
			Flags:    wantFlags,
			Length:   0x8,
			StreamID: 0,
		},
		Data: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	if !reflect.DeepEqual(f, want) {
		t.Errorf("parsed back %#v; want %#v", f, want)
	}
}

func TestReadFrameHeader(t *testing.T) {
	tests := []struct {
		in   string
		want http2FrameHeader
	}{
		{in: "\x00\x00\x00" + "\x00" + "\x00" + "\x00\x00\x00\x00", want: http2FrameHeader{}},
		{in: "\x01\x02\x03" + "\x04" + "\x05" + "\x06\x07\x08\x09", want: http2FrameHeader{
			Length: 66051, Type: 4, Flags: 5, StreamID: 101124105,
		}},
		// Ignore high bit:
		{in: "\xff\xff\xff" + "\xff" + "\xff" + "\xff\xff\xff\xff", want: http2FrameHeader{
			Length: 16777215, Type: 255, Flags: 255, StreamID: 2147483647}},
		{in: "\xff\xff\xff" + "\xff" + "\xff" + "\x7f\xff\xff\xff", want: http2FrameHeader{
			Length: 16777215, Type: 255, Flags: 255, StreamID: 2147483647}},
	}
	for i, tt := range tests {
		got, err := http2readFrameHeader(make([]byte, 9), strings.NewReader(tt.in))
		if err != nil {
			t.Errorf("%d. readFrameHeader(%q) = %v", i, tt.in, err)
			continue
		}
		tt.want.valid = true
		if !got.Equal(tt.want) {
			t.Errorf("%d. readFrameHeader(%q) = %+v; want %+v", i, tt.in, got, tt.want)
		}
	}
}

func TestReadWriteFrameHeader(t *testing.T) {
	tests := []struct {
		len      uint32
		typ      http2FrameType
		flags    http2Flags
		streamID uint32
	}{
		{len: 0, typ: 255, flags: 1, streamID: 0},
		{len: 0, typ: 255, flags: 1, streamID: 1},
		{len: 0, typ: 255, flags: 1, streamID: 255},
		{len: 0, typ: 255, flags: 1, streamID: 256},
		{len: 0, typ: 255, flags: 1, streamID: 65535},
		{len: 0, typ: 255, flags: 1, streamID: 65536},

		{len: 0, typ: 1, flags: 255, streamID: 1},
		{len: 255, typ: 1, flags: 255, streamID: 1},
		{len: 256, typ: 1, flags: 255, streamID: 1},
		{len: 65535, typ: 1, flags: 255, streamID: 1},
		{len: 65536, typ: 1, flags: 255, streamID: 1},
		{len: 16777215, typ: 1, flags: 255, streamID: 1},
	}
	for _, tt := range tests {
		fr, buf := testFramer()
		fr.startWrite(tt.typ, tt.flags, tt.streamID)
		fr.writeBytes(make([]byte, tt.len))
		fr.endWrite()
		fh, err := http2ReadFrameHeader(buf)
		if err != nil {
			t.Errorf("ReadFrameHeader(%+v) = %v", tt, err)
			continue
		}
		if fh.Type != tt.typ || fh.Flags != tt.flags || fh.Length != tt.len || fh.StreamID != tt.streamID {
			t.Errorf("ReadFrameHeader(%+v) = %+v; mismatch", tt, fh)
		}
	}

}

func TestWriteTooLargeFrame(t *testing.T) {
	fr, _ := testFramer()
	fr.startWrite(0, 1, 1)
	fr.writeBytes(make([]byte, 1<<24))
	err := fr.endWrite()
	if err != http2ErrFrameTooLarge {
		t.Errorf("endWrite = %v; want errFrameTooLarge", err)
	}
}

func TestWriteGoAway(t *testing.T) {
	const debug = "foo"
	fr, buf := testFramer()
	if err := fr.WriteGoAway(0x01020304, 0x05060708, []byte(debug)); err != nil {
		t.Fatal(err)
	}
	const wantEnc = "\x00\x00\v\a\x00\x00\x00\x00\x00\x01\x02\x03\x04\x05\x06\x07\x08" + debug
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	want := &http2GoAwayFrame{
		http2FrameHeader: http2FrameHeader{
			valid:    true,
			Type:     0x7,
			Flags:    0,
			Length:   uint32(4 + 4 + len(debug)),
			StreamID: 0,
		},
		LastStreamID: 0x01020304,
		ErrCode:      0x05060708,
		debugData:    []byte(debug),
	}
	if !reflect.DeepEqual(f, want) {
		t.Fatalf("parsed back:\n%#v\nwant:\n%#v", f, want)
	}
	if got := string(f.(*http2GoAwayFrame).DebugData()); got != debug {
		t.Errorf("debug data = %q; want %q", got, debug)
	}
}

func TestWritePushPromise(t *testing.T) {
	pp := http2PushPromiseParam{
		StreamID:      42,
		PromiseID:     42,
		BlockFragment: []byte("abc"),
	}
	fr, buf := testFramer()
	if err := fr.WritePushPromise(pp); err != nil {
		t.Fatal(err)
	}
	const wantEnc = "\x00\x00\x07\x05\x00\x00\x00\x00*\x00\x00\x00*abc"
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	_, ok := f.(*http2PushPromiseFrame)
	if !ok {
		t.Fatalf("got %T; want *PushPromiseFrame", f)
	}
	want := &http2PushPromiseFrame{
		http2FrameHeader: http2FrameHeader{
			valid:    true,
			Type:     0x5,
			Flags:    0x0,
			Length:   0x7,
			StreamID: 42,
		},
		PromiseID:     42,
		headerFragBuf: []byte("abc"),
	}
	if !reflect.DeepEqual(f, want) {
		t.Fatalf("parsed back:\n%#v\nwant:\n%#v", f, want)
	}
}

// test checkFrameOrder and that HEADERS and CONTINUATION frames can't be intermingled.
func TestReadFrameOrder(t *testing.T) {
	head := func(f *http2Framer, id uint32, end bool) {
		f.WriteHeaders(http2HeadersFrameParam{
			StreamID:      id,
			BlockFragment: []byte("foo"), // unused, but non-empty
			EndHeaders:    end,
		})
	}
	cont := func(f *http2Framer, id uint32, end bool) {
		f.WriteContinuation(id, end, []byte("foo"))
	}

	tests := [...]struct {
		name    string
		w       func(*http2Framer)
		atLeast int
		wantErr string
	}{
		0: {
			w: func(f *http2Framer) {
				head(f, 1, true)
			},
		},
		1: {
			w: func(f *http2Framer) {
				head(f, 1, true)
				head(f, 2, true)
			},
		},
		2: {
			wantErr: "got HEADERS for stream 2; expected CONTINUATION following HEADERS for stream 1",
			w: func(f *http2Framer) {
				head(f, 1, false)
				head(f, 2, true)
			},
		},
		3: {
			wantErr: "got DATA for stream 1; expected CONTINUATION following HEADERS for stream 1",
			w: func(f *http2Framer) {
				head(f, 1, false)
			},
		},
		4: {
			w: func(f *http2Framer) {
				head(f, 1, false)
				cont(f, 1, true)
				head(f, 2, true)
			},
		},
		5: {
			wantErr: "got CONTINUATION for stream 2; expected stream 1",
			w: func(f *http2Framer) {
				head(f, 1, false)
				cont(f, 2, true)
				head(f, 2, true)
			},
		},
		6: {
			wantErr: "unexpected CONTINUATION for stream 1",
			w: func(f *http2Framer) {
				cont(f, 1, true)
			},
		},
		7: {
			wantErr: "unexpected CONTINUATION for stream 1",
			w: func(f *http2Framer) {
				cont(f, 1, false)
			},
		},
		8: {
			wantErr: "HEADERS frame with stream ID 0",
			w: func(f *http2Framer) {
				head(f, 0, true)
			},
		},
		9: {
			wantErr: "CONTINUATION frame with stream ID 0",
			w: func(f *http2Framer) {
				cont(f, 0, true)
			},
		},
		10: {
			wantErr: "unexpected CONTINUATION for stream 1",
			atLeast: 5,
			w: func(f *http2Framer) {
				head(f, 1, false)
				cont(f, 1, false)
				cont(f, 1, false)
				cont(f, 1, false)
				cont(f, 1, true)
				cont(f, 1, false)
			},
		},
	}
	for i, tt := range tests {
		buf := new(bytes.Buffer)
		f := http2NewFramer(buf, buf)
		f.AllowIllegalWrites = true
		tt.w(f)
		f.WriteData(1, true, nil) // to test transition away from last step

		var err error
		n := 0
		var log bytes.Buffer
		for {
			var got http2Frame
			got, err = f.ReadFrame()
			fmt.Fprintf(&log, "  read %v, %v\n", got, err)
			if err != nil {
				break
			}
			n++
		}
		if err == io.EOF {
			err = nil
		}
		ok := tt.wantErr == ""
		if ok && err != nil {
			t.Errorf("%d. after %d good frames, ReadFrame = %v; want success\n%s", i, n, err, log.Bytes())
			continue
		}
		if !ok && err != http2ConnectionError(http2ErrCodeProtocol) {
			t.Errorf("%d. after %d good frames, ReadFrame = %v; want ConnectionError(ErrCodeProtocol)\n%s", i, n, err, log.Bytes())
			continue
		}
		if !((f.errDetail == nil && tt.wantErr == "") || (fmt.Sprint(f.errDetail) == tt.wantErr)) {
			t.Errorf("%d. framer eror = %q; want %q\n%s", i, f.errDetail, tt.wantErr, log.Bytes())
		}
		if n < tt.atLeast {
			t.Errorf("%d. framer only read %d frames; want at least %d\n%s", i, n, tt.atLeast, log.Bytes())
		}
	}
}

type hpackEncoder struct {
	enc *hpack.Encoder
	buf bytes.Buffer
}

func (he *hpackEncoder) encodeHeaderRaw(t *testing.T, headers ...string) []byte {
	if len(headers)%2 == 1 {
		panic("odd number of kv args")
	}
	he.buf.Reset()
	if he.enc == nil {
		he.enc = hpack.NewEncoder(&he.buf)
	}
	for len(headers) > 0 {
		k, v := headers[0], headers[1]
		err := he.enc.WriteField(hpack.HeaderField{Name: k, Value: v})
		if err != nil {
			t.Fatalf("HPACK encoding error for %q/%q: %v", k, v, err)
		}
		headers = headers[2:]
	}
	return he.buf.Bytes()
}

func TestMetaFrameHeader(t *testing.T) {
	write := func(f *http2Framer, frags ...[]byte) {
		for i, frag := range frags {
			end := (i == len(frags)-1)
			if i == 0 {
				f.WriteHeaders(http2HeadersFrameParam{
					StreamID:      1,
					BlockFragment: frag,
					EndHeaders:    end,
				})
			} else {
				f.WriteContinuation(1, end, frag)
			}
		}
	}

	want := func(flags http2Flags, length uint32, pairs ...string) *http2MetaHeadersFrame {
		mh := &http2MetaHeadersFrame{
			http2HeadersFrame: &http2HeadersFrame{
				http2FrameHeader: http2FrameHeader{
					Type:     http2FrameHeaders,
					Flags:    flags,
					Length:   length,
					StreamID: 1,
				},
			},
			Fields: []hpack.HeaderField(nil),
		}
		for len(pairs) > 0 {
			mh.Fields = append(mh.Fields, hpack.HeaderField{
				Name:  pairs[0],
				Value: pairs[1],
			})
			pairs = pairs[2:]
		}
		return mh
	}
	truncated := func(mh *http2MetaHeadersFrame) *http2MetaHeadersFrame {
		mh.Truncated = true
		return mh
	}

	const noFlags http2Flags = 0

	oneKBString := strings.Repeat("a", 1<<10)

	tests := [...]struct {
		name              string
		w                 func(*http2Framer)
		want              interface{} // *MetaHeaderFrame or error
		wantErrReason     string
		maxHeaderListSize uint32
	}{
		0: {
			name: "single_headers",
			w: func(f *http2Framer) {
				var he hpackEncoder
				all := he.encodeHeaderRaw(t, ":method", "GET", ":path", "/")
				write(f, all)
			},
			want: want(http2FlagHeadersEndHeaders, 2, ":method", "GET", ":path", "/"),
		},
		1: {
			name: "with_continuation",
			w: func(f *http2Framer) {
				var he hpackEncoder
				all := he.encodeHeaderRaw(t, ":method", "GET", ":path", "/", "foo", "bar")
				write(f, all[:1], all[1:])
			},
			want: want(noFlags, 1, ":method", "GET", ":path", "/", "foo", "bar"),
		},
		2: {
			name: "with_two_continuation",
			w: func(f *http2Framer) {
				var he hpackEncoder
				all := he.encodeHeaderRaw(t, ":method", "GET", ":path", "/", "foo", "bar")
				write(f, all[:2], all[2:4], all[4:])
			},
			want: want(noFlags, 2, ":method", "GET", ":path", "/", "foo", "bar"),
		},
		3: {
			name: "big_string_okay",
			w: func(f *http2Framer) {
				var he hpackEncoder
				all := he.encodeHeaderRaw(t, ":method", "GET", ":path", "/", "foo", oneKBString)
				write(f, all[:2], all[2:])
			},
			want: want(noFlags, 2, ":method", "GET", ":path", "/", "foo", oneKBString),
		},
		4: {
			name: "big_string_error",
			w: func(f *http2Framer) {
				var he hpackEncoder
				all := he.encodeHeaderRaw(t, ":method", "GET", ":path", "/", "foo", oneKBString)
				write(f, all[:2], all[2:])
			},
			maxHeaderListSize: (1 << 10) / 2,
			want:              http2ConnectionError(http2ErrCodeCompression),
		},
		5: {
			name: "max_header_list_truncated",
			w: func(f *http2Framer) {
				var he hpackEncoder
				var pairs = []string{":method", "GET", ":path", "/"}
				for i := 0; i < 100; i++ {
					pairs = append(pairs, "foo", "bar")
				}
				all := he.encodeHeaderRaw(t, pairs...)
				write(f, all[:2], all[2:])
			},
			maxHeaderListSize: (1 << 10) / 2,
			want: truncated(want(noFlags, 2,
				":method", "GET",
				":path", "/",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar",
				"foo", "bar", // 11
			)),
		},
		6: {
			name: "pseudo_order",
			w: func(f *http2Framer) {
				write(f, encodeHeaderRaw(t,
					":method", "GET",
					"foo", "bar",
					":path", "/", // bogus
				))
			},
			want:          http2streamError(1, http2ErrCodeProtocol),
			wantErrReason: "pseudo header field after regular",
		},
		7: {
			name: "pseudo_unknown",
			w: func(f *http2Framer) {
				write(f, encodeHeaderRaw(t,
					":unknown", "foo", // bogus
					"foo", "bar",
				))
			},
			want:          http2streamError(1, http2ErrCodeProtocol),
			wantErrReason: "invalid pseudo-header \":unknown\"",
		},
		8: {
			name: "pseudo_mix_request_response",
			w: func(f *http2Framer) {
				write(f, encodeHeaderRaw(t,
					":method", "GET",
					":status", "100",
				))
			},
			want:          http2streamError(1, http2ErrCodeProtocol),
			wantErrReason: "mix of request and response pseudo headers",
		},
		9: {
			name: "pseudo_dup",
			w: func(f *http2Framer) {
				write(f, encodeHeaderRaw(t,
					":method", "GET",
					":method", "POST",
				))
			},
			want:          http2streamError(1, http2ErrCodeProtocol),
			wantErrReason: "duplicate pseudo-header \":method\"",
		},
		10: {
			name: "trailer_okay_no_pseudo",
			w:    func(f *http2Framer) { write(f, encodeHeaderRaw(t, "foo", "bar")) },
			want: want(http2FlagHeadersEndHeaders, 8, "foo", "bar"),
		},
		11: {
			name:          "invalid_field_name",
			w:             func(f *http2Framer) { write(f, encodeHeaderRaw(t, "CapitalBad", "x")) },
			want:          http2streamError(1, http2ErrCodeProtocol),
			wantErrReason: "invalid header field name \"CapitalBad\"",
		},
		12: {
			name:          "invalid_field_value",
			w:             func(f *http2Framer) { write(f, encodeHeaderRaw(t, "key", "bad_null\x00")) },
			want:          http2streamError(1, http2ErrCodeProtocol),
			wantErrReason: "invalid header field value \"bad_null\\x00\"",
		},
	}
	for i, tt := range tests {
		buf := new(bytes.Buffer)
		f := http2NewFramer(buf, buf)
		f.ReadMetaHeaders = hpack.NewDecoder(http2initialHeaderTableSize, nil)
		f.MaxHeaderListSize = tt.maxHeaderListSize
		tt.w(f)

		name := tt.name
		if name == "" {
			name = fmt.Sprintf("test index %d", i)
		}

		var got interface{}
		var err error
		got, err = f.ReadFrame()
		if err != nil {
			got = err

			// Ignore the StreamError.Cause field, if it matches the wantErrReason.
			// The test table above predates the Cause field.
			if se, ok := err.(http2StreamError); ok && se.Cause != nil && se.Cause.Error() == tt.wantErrReason {
				se.Cause = nil
				got = se
			}
		}
		if !reflect.DeepEqual(got, tt.want) {
			if mhg, ok := got.(*http2MetaHeadersFrame); ok {
				if mhw, ok := tt.want.(*http2MetaHeadersFrame); ok {
					hg := mhg.http2HeadersFrame
					hw := mhw.http2HeadersFrame
					if hg != nil && hw != nil && !reflect.DeepEqual(*hg, *hw) {
						t.Errorf("%s: headers differ:\n got: %+v\nwant: %+v\n", name, *hg, *hw)
					}
				}
			}
			str := func(v interface{}) string {
				if _, ok := v.(error); ok {
					return fmt.Sprintf("error %v", v)
				} else {
					return fmt.Sprintf("value %#v", v)
				}
			}
			t.Errorf("%s:\n got: %v\nwant: %s", name, str(got), str(tt.want))
		}
		if tt.wantErrReason != "" && tt.wantErrReason != fmt.Sprint(f.errDetail) {
			t.Errorf("%s: got error reason %q; want %q", name, f.errDetail, tt.wantErrReason)
		}
	}
}

func TestSetReuseFrames(t *testing.T) {
	fr, buf := testFramer()
	fr.SetReuseFrames()

	// Check that DataFrames are reused. Note that
	// SetReuseFrames only currently implements reuse of DataFrames.
	firstDf := readAndVerifyDataFrame("ABC", 3, fr, buf, t)

	for i := 0; i < 10; i++ {
		df := readAndVerifyDataFrame("XYZ", 3, fr, buf, t)
		if df != firstDf {
			t.Errorf("Expected Framer to return references to the same DataFrame. Have %v and %v", &df, &firstDf)
		}
	}

	for i := 0; i < 10; i++ {
		df := readAndVerifyDataFrame("", 0, fr, buf, t)
		if df != firstDf {
			t.Errorf("Expected Framer to return references to the same DataFrame. Have %v and %v", &df, &firstDf)
		}
	}

	for i := 0; i < 10; i++ {
		df := readAndVerifyDataFrame("HHH", 3, fr, buf, t)
		if df != firstDf {
			t.Errorf("Expected Framer to return references to the same DataFrame. Have %v and %v", &df, &firstDf)
		}
	}
}

func TestSetReuseFramesMoreThanOnce(t *testing.T) {
	fr, buf := testFramer()
	fr.SetReuseFrames()

	firstDf := readAndVerifyDataFrame("ABC", 3, fr, buf, t)
	fr.SetReuseFrames()

	for i := 0; i < 10; i++ {
		df := readAndVerifyDataFrame("XYZ", 3, fr, buf, t)
		// SetReuseFrames should be idempotent
		fr.SetReuseFrames()
		if df != firstDf {
			t.Errorf("Expected Framer to return references to the same DataFrame. Have %v and %v", &df, &firstDf)
		}
	}
}

func TestNoSetReuseFrames(t *testing.T) {
	fr, buf := testFramer()
	const numNewDataFrames = 10
	dfSoFar := make([]interface{}, numNewDataFrames)

	// Check that DataFrames are not reused if SetReuseFrames wasn't called.
	// SetReuseFrames only currently implements reuse of DataFrames.
	for i := 0; i < numNewDataFrames; i++ {
		df := readAndVerifyDataFrame("XYZ", 3, fr, buf, t)
		for _, item := range dfSoFar {
			if df == item {
				t.Errorf("Expected Framer to return new DataFrames since SetNoReuseFrames not set.")
			}
		}
		dfSoFar[i] = df
	}
}

func readAndVerifyDataFrame(data string, length byte, fr *http2Framer, buf *bytes.Buffer, t *testing.T) *http2DataFrame {
	var streamID uint32 = 1<<24 + 2<<16 + 3<<8 + 4
	fr.WriteData(streamID, true, []byte(data))
	wantEnc := "\x00\x00" + string(length) + "\x00\x01\x01\x02\x03\x04" + data
	if buf.String() != wantEnc {
		t.Errorf("encoded as %q; want %q", buf.Bytes(), wantEnc)
	}
	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}
	df, ok := f.(*http2DataFrame)
	if !ok {
		t.Fatalf("got %T; want *http2DataFrame", f)
	}
	if !bytes.Equal(df.Data(), []byte(data)) {
		t.Errorf("got %q; want %q", df.Data(), []byte(data))
	}
	if f.Header().Flags&1 == 0 {
		t.Errorf("didn't see END_STREAM flag")
	}
	return df
}

func encodeHeaderRaw(t *testing.T, pairs ...string) []byte {
	var he hpackEncoder
	return he.encodeHeaderRaw(t, pairs...)
}

func TestSettingsDuplicates(t *testing.T) {
	tests := []struct {
		settings []http2Setting
		want     bool
	}{
		{nil, false},
		{[]http2Setting{{ID: 1}}, false},
		{[]http2Setting{{ID: 1}, {ID: 2}}, false},
		{[]http2Setting{{ID: 1}, {ID: 2}}, false},
		{[]http2Setting{{ID: 1}, {ID: 2}, {ID: 3}}, false},
		{[]http2Setting{{ID: 1}, {ID: 2}, {ID: 3}}, false},
		{[]http2Setting{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}, false},

		{[]http2Setting{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 2}}, true},
		{[]http2Setting{{ID: 4}, {ID: 2}, {ID: 3}, {ID: 4}}, true},

		{[]http2Setting{
			{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4},
			{ID: 5}, {ID: 6}, {ID: 7}, {ID: 8},
			{ID: 9}, {ID: 10}, {ID: 11}, {ID: 12},
		}, false},

		{[]http2Setting{
			{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4},
			{ID: 5}, {ID: 6}, {ID: 7}, {ID: 8},
			{ID: 9}, {ID: 10}, {ID: 11}, {ID: 11},
		}, true},
	}
	for i, tt := range tests {
		fr, _ := testFramer()
		fr.WriteSettings(tt.settings...)
		f, err := fr.ReadFrame()
		if err != nil {
			t.Fatalf("%d. ReadFrame: %v", i, err)
		}
		sf := f.(*http2SettingsFrame)
		got := sf.HasDuplicates()
		if got != tt.want {
			t.Errorf("%d. HasDuplicates = %v; want %v", i, got, tt.want)
		}
	}

}

func TestParseSettingsFrame(t *testing.T) {
	fh := http2FrameHeader{}
	fh.Flags = http2FlagSettingsAck
	fh.Length = 1
	countErr := func(s string) {}
	_, err := http2parseSettingsFrame(nil, fh, countErr, nil)
	tests.AssertErrorContains(t, err, "FRAME_SIZE_ERROR")

	fh = http2FrameHeader{StreamID: 1}
	_, err = http2parseSettingsFrame(nil, fh, countErr, nil)
	tests.AssertErrorContains(t, err, "PROTOCOL_ERROR")

	fh = http2FrameHeader{}
	_, err = http2parseSettingsFrame(nil, fh, countErr, []byte("roc"))
	tests.AssertErrorContains(t, err, "FRAME_SIZE_ERROR")

	fh = http2FrameHeader{valid: true}
	_, err = http2parseSettingsFrame(nil, fh, countErr, []byte("rocroc"))
	tests.AssertNoError(t, err)
}

func TestParsePushPromise(t *testing.T) {
	fh := http2FrameHeader{}
	countError := func(string) {}
	_, err := http2parsePushPromise(nil, fh, countError, nil)
	tests.AssertErrorContains(t, err, "PROTOCOL_ERROR")

	fh.StreamID = 1
	fh.Flags = http2FlagPushPromisePadded
	_, err = http2parsePushPromise(nil, fh, countError, nil)
	tests.AssertErrorContains(t, err, "EOF")

	fh.Flags = 0
	_, err = http2parsePushPromise(nil, fh, countError, nil)
	tests.AssertErrorContains(t, err, "EOF")

	_, err = http2parsePushPromise(nil, fh, countError, []byte("ksjfksjksjflskk"))
	tests.AssertNoError(t, err)
}

func TestSummarizeFrame(t *testing.T) {
	fh := http2FrameHeader{valid: true}
	var f http2Frame
	f = &http2SettingsFrame{http2FrameHeader: fh, p: []byte{0x09, 0x01, 0x80, 0x20, 0x00, 0x11}}
	s := http2summarizeFrame(f)
	tests.AssertContains(t, s, "len=0", true)

	f = &http2DataFrame{http2FrameHeader: fh}
	s = http2summarizeFrame(f)
	tests.AssertContains(t, s, `data=""`, true)

	f = &http2WindowUpdateFrame{http2FrameHeader: fh}
	s = http2summarizeFrame(f)
	tests.AssertContains(t, s, "conn", true)

	f = &http2PingFrame{http2FrameHeader: fh}
	s = http2summarizeFrame(f)
	tests.AssertContains(t, s, "ping", true)

	f = &http2GoAwayFrame{http2FrameHeader: fh}
	s = http2summarizeFrame(f)
	tests.AssertContains(t, s, "laststreamid", true)

	f = &http2RSTStreamFrame{http2FrameHeader: fh}
	s = http2summarizeFrame(f)
	tests.AssertContains(t, s, "no_error", true)
}

func TestParseDataFrame(t *testing.T) {
	fh := http2FrameHeader{valid: true}
	countError := func(string) {}
	_, err := http2parseDataFrame(nil, fh, countError, nil)
	tests.AssertErrorContains(t, err, "DATA frame with stream ID 0")

	fh.StreamID = 1
	fh.Flags = http2FlagDataPadded
	fc := &http2frameCache{}
	payload := []byte{0x09, 0x00, 0x00, 0x98, 0x11, 0x12}
	_, err = http2parseDataFrame(fc, fh, countError, payload)
	tests.AssertErrorContains(t, err, "pad size larger than data payload")

	payload = []byte{0x02, 0x00, 0x00, 0x98, 0x11, 0x12}
	_, err = http2parseDataFrame(fc, fh, countError, payload)
	tests.AssertNoError(t, err)
}

func TestParseWindowUpdateFrame(t *testing.T) {
	fh := http2FrameHeader{valid: true}
	countError := func(string) {}
	_, err := http2parseWindowUpdateFrame(nil, fh, countError, nil)
	tests.AssertErrorContains(t, err, "FRAME_SIZE_ERROR")

	p := []byte{0x00, 0x00, 0x00, 0x00}
	_, err = http2parseWindowUpdateFrame(nil, fh, countError, p)
	tests.AssertErrorContains(t, err, "PROTOCOL_ERROR")

	fh.StreamID = 255
	p[0] = 0x01
	p[3] = 0x01
	_, err = http2parseWindowUpdateFrame(nil, fh, countError, p)
	tests.AssertNoError(t, err)
}

func TestParseUnknownFrame(t *testing.T) {
	fh := http2FrameHeader{valid: true}
	countError := func(string) {}
	p := []byte("test")
	f, err := http2parseUnknownFrame(nil, fh, countError, p)
	tests.AssertNoError(t, err)
	uf, ok := f.(*http2UnknownFrame)
	if !ok {
		t.Fatalf("not http2UnknownFrame type: %#+v", f)
	}
	assertEqual(t, p, uf.Payload())
}

func TestParseRSTStreamFrame(t *testing.T) {
	fh := http2FrameHeader{valid: true}
	countError := func(string) {}
	p := []byte("test.")
	_, err := http2parseRSTStreamFrame(nil, fh, countError, p)
	tests.AssertErrorContains(t, err, "FRAME_SIZE_ERROR")

	p = []byte("test")
	_, err = http2parseRSTStreamFrame(nil, fh, countError, p)
	tests.AssertErrorContains(t, err, "PROTOCOL_ERROR")

	fh.StreamID = 1
	_, err = http2parseRSTStreamFrame(nil, fh, countError, p)
	tests.AssertNoError(t, err)
}

func TestParsePingFrame(t *testing.T) {
	fh := http2FrameHeader{valid: true}
	countError := func(string) {}
	payload := []byte("")
	_, err := http2parsePingFrame(nil, fh, countError, payload)
	tests.AssertErrorContains(t, err, "FRAME_SIZE_ERROR")

	payload = []byte("testtest")
	fh.StreamID = 1
	_, err = http2parsePingFrame(nil, fh, countError, payload)
	tests.AssertErrorContains(t, err, "PROTOCOL_ERROR")

	fh.StreamID = 0
	_, err = http2parsePingFrame(nil, fh, countError, payload)
	tests.AssertNoError(t, err)
}

func TestPushPromiseFrame(t *testing.T) {
	fh := http2FrameHeader{valid: true}
	buf := []byte("test")
	f := &http2PushPromiseFrame{http2FrameHeader: fh, headerFragBuf: buf}
	assertEqual(t, buf, f.HeaderBlockFragment())
	assertEqual(t, false, f.HeadersEnded())
}

func TestH2Framer(t *testing.T) {
	f := &http2Framer{}
	f.debugWriteLoggerf = func(s string, i ...interface{}) {}
	f.logWrite()
	assertNotNil(t, f.debugFramer)
	assertNil(t, f.ErrorDetail())

	f.w = new(bytes.Buffer)
	err := f.WriteRawFrame(http2FrameData, http2FlagDataEndStream, 1, nil)
	tests.AssertNoError(t, err)

	param := http2PushPromiseParam{}
	err = f.WritePushPromise(param)
	tests.AssertErrorContains(t, err, "invalid stream ID")

	param.StreamID = 1
	param.EndHeaders = true
	param.PadLength = 2
	f.AllowIllegalWrites = true
	err = f.WritePushPromise(param)
	tests.AssertNoError(t, err)
}
