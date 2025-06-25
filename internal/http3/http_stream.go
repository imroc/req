package http3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"

	"github.com/imroc/req/v3/internal/compress"
	"github.com/imroc/req/v3/internal/dump"
	"github.com/imroc/req/v3/internal/transport"

	"github.com/quic-go/quic-go"

	"github.com/quic-go/qpack"
)

// A stream is an HTTP/3 request stream.
// When writing to and reading from the stream, data is framed in HTTP/3 DATA frames.
type stream struct {
	*quic.Stream

	conn *connection

	buf []byte // used as a temporary buffer when writing the HTTP/3 frame headers

	bytesRemainingInFrame uint64

	datagrams *datagrammer

	parseTrailer  func(io.Reader, uint64) error
	parsedTrailer bool
}

func newStream(str *quic.Stream, conn *connection, datagrams *datagrammer, parseTrailer func(io.Reader, uint64) error) *stream {
	return &stream{
		Stream:       str,
		conn:         conn,
		buf:          make([]byte, 16),
		datagrams:    datagrams,
		parseTrailer: parseTrailer,
	}
}

func (s *stream) Read(b []byte) (int, error) {
	fp := &frameParser{
		r:    s.Stream,
		conn: s.conn.Conn,
	}
	if s.bytesRemainingInFrame == 0 {
	parseLoop:
		for {
			frame, err := fp.ParseNext()
			if err != nil {
				return 0, err
			}
			switch f := frame.(type) {
			case *dataFrame:
				if s.parsedTrailer {
					return 0, errors.New("DATA frame received after trailers")
				}
				s.bytesRemainingInFrame = f.Length
				break parseLoop
			case *headersFrame:
				if s.conn.perspective == PerspectiveServer {
					continue
				}
				if s.parsedTrailer {
					return 0, errors.New("additional HEADERS frame received after trailers")
				}
				s.parsedTrailer = true
				return 0, s.parseTrailer(s.Stream, f.Length)
			default:
				s.conn.CloseWithError(quic.ApplicationErrorCode(ErrCodeFrameUnexpected), "")
				// parseNextFrame skips over unknown frame types
				// Therefore, this condition is only entered when we parsed another known frame type.
				return 0, fmt.Errorf("peer sent an unexpected frame: %T", f)
			}
		}
	}

	var n int
	var err error
	if s.bytesRemainingInFrame < uint64(len(b)) {
		n, err = s.Stream.Read(b[:s.bytesRemainingInFrame])
	} else {
		n, err = s.Stream.Read(b)
	}
	s.bytesRemainingInFrame -= uint64(n)
	return n, err
}

func (s *stream) hasMoreData() bool {
	return s.bytesRemainingInFrame > 0
}

func (s *stream) Write(b []byte) (int, error) {
	s.buf = s.buf[:0]
	s.buf = (&dataFrame{Length: uint64(len(b))}).Append(s.buf)
	if _, err := s.Stream.Write(s.buf); err != nil {
		return 0, err
	}
	return s.Stream.Write(b)
}

func (s *stream) writeUnframed(b []byte) (int, error) {
	return s.Stream.Write(b)
}

func (s *stream) StreamID() quic.StreamID {
	return s.Stream.StreamID()
}

// The stream conforms to the quic.Stream interface, but instead of writing to and reading directly
// from the QUIC stream, it writes to and reads from the HTTP stream.
type requestStream struct {
	ctx context.Context
	*stream
	*transport.Options

	responseBody io.ReadCloser // set by ReadResponse

	decoder            *qpack.Decoder
	requestWriter      *requestWriter
	maxHeaderBytes     uint64
	reqDone            chan<- struct{}
	disableCompression bool
	response           *http.Response
	trace              *httptrace.ClientTrace

	sentRequest   bool
	requestedGzip bool
	isConnect     bool
	firstByte     bool
}

func newRequestStream(
	options *transport.Options,
	str *stream,
	requestWriter *requestWriter,
	reqDone chan<- struct{},
	decoder *qpack.Decoder,
	disableCompression bool,
	maxHeaderBytes uint64,
	rsp *http.Response,
	trace *httptrace.ClientTrace,
) *requestStream {
	return &requestStream{
		Options:            options,
		stream:             str,
		requestWriter:      requestWriter,
		reqDone:            reqDone,
		decoder:            decoder,
		disableCompression: disableCompression,
		maxHeaderBytes:     maxHeaderBytes,
		response:           rsp,
		trace:              trace,
	}
}

func (s *requestStream) Read(b []byte) (int, error) {
	if s.responseBody == nil {
		return 0, errors.New("http3: invalid use of requestStream.Read: need to call ReadResponse first")
	}
	return s.responseBody.Read(b)
}

func (s *requestStream) SendRequestHeader(req *http.Request) error {
	if s.sentRequest {
		return errors.New("http3: invalid duplicate use of SendRequestHeader")
	}
	if !s.DisableCompression && !s.disableCompression && req.Method != http.MethodHead &&
		req.Header.Get("Accept-Encoding") == "" && req.Header.Get("Range") == "" {
		s.requestedGzip = true
	}
	dumps := dump.GetDumpers(req.Context(), s.Dump)
	var headerDumps []*dump.Dumper
	for _, dump := range dumps {
		if dump.RequestHeader() {
			headerDumps = append(headerDumps, dump)
		}
	}

	s.isConnect = req.Method == http.MethodConnect
	s.sentRequest = true
	return s.requestWriter.WriteRequestHeader(s.Stream, req, s.requestedGzip, headerDumps)
}

func (s *requestStream) ReadResponse() (*http.Response, error) {
	fp := &frameParser{
		conn: s.conn.Conn,
		r: &tracingReader{
			Reader: s.Stream,
			first:  &s.firstByte,
			trace:  s.trace,
		},
	}
	frame, err := fp.ParseNext()
	if err != nil {
		s.CancelRead(quic.StreamErrorCode(ErrCodeFrameError))
		s.CancelWrite(quic.StreamErrorCode(ErrCodeFrameError))
		return nil, fmt.Errorf("http3: parsing frame failed: %w", err)
	}
	hf, ok := frame.(*headersFrame)
	if !ok {
		s.conn.CloseWithError(quic.ApplicationErrorCode(ErrCodeFrameUnexpected), "expected first frame to be a HEADERS frame")
		return nil, errors.New("http3: expected first frame to be a HEADERS frame")
	}
	if hf.Length > s.maxHeaderBytes {
		s.CancelRead(quic.StreamErrorCode(ErrCodeFrameError))
		s.CancelWrite(quic.StreamErrorCode(ErrCodeFrameError))
		return nil, fmt.Errorf("http3: HEADERS frame too large: %d bytes (max: %d)", hf.Length, s.maxHeaderBytes)
	}
	headerBlock := make([]byte, hf.Length)
	if _, err := io.ReadFull(s.Stream, headerBlock); err != nil {
		s.CancelRead(quic.StreamErrorCode(ErrCodeRequestIncomplete))
		s.CancelWrite(quic.StreamErrorCode(ErrCodeRequestIncomplete))
		return nil, fmt.Errorf("http3: failed to read response headers: %w", err)
	}
	hfs, err := s.decoder.DecodeFull(headerBlock)
	if err != nil {
		// TODO: use the right error code
		s.conn.CloseWithError(quic.ApplicationErrorCode(ErrCodeGeneralProtocolError), "")
		return nil, fmt.Errorf("http3: failed to decode response headers: %w", err)
	}
	ds := dump.GetResponseHeaderDumpers(s.ctx, s.Dump)
	if ds.ShouldDump() {
		for _, h := range hfs {
			ds.DumpResponseHeader([]byte(fmt.Sprintf("%s: %s\r\n", h.Name, h.Value)))
		}
		ds.DumpResponseHeader([]byte("\r\n"))
	}
	res := s.response
	if err := updateResponseFromHeaders(res, hfs); err != nil {
		s.CancelRead(quic.StreamErrorCode(ErrCodeMessageError))
		s.CancelWrite(quic.StreamErrorCode(ErrCodeMessageError))
		return nil, fmt.Errorf("http3: invalid response: %w", err)
	}

	// Check that the server doesn't send more data in DATA frames than indicated by the Content-Length header (if set).
	// See section 4.1.2 of RFC 9114.
	respBody := newResponseBody(s.stream, res.ContentLength, s.reqDone)

	// Rules for when to set Content-Length are defined in https://tools.ietf.org/html/rfc7230#section-3.3.2.
	isInformational := res.StatusCode >= 100 && res.StatusCode < 200
	isNoContent := res.StatusCode == http.StatusNoContent
	isSuccessfulConnect := s.isConnect && res.StatusCode >= 200 && res.StatusCode < 300
	if (isInformational || isNoContent || isSuccessfulConnect) && res.ContentLength == -1 {
		res.ContentLength = 0
	}
	if s.requestedGzip && res.Header.Get("Content-Encoding") == "gzip" {
		res.Header.Del("Content-Encoding")
		res.Header.Del("Content-Length")
		res.ContentLength = -1
		s.responseBody = compress.NewGzipReader(respBody)
		res.Uncompressed = true
	} else if s.AutoDecompression {
		contentEncoding := res.Header.Get("Content-Encoding")
		if contentEncoding != "" {
			res.Header.Del("Content-Encoding")
			res.Header.Del("Content-Length")
			res.ContentLength = -1
			res.Uncompressed = true
			res.Body = compress.NewCompressReader(respBody, contentEncoding)
		}
	} else {
		s.responseBody = respBody
	}
	res.Body = s.responseBody
	return res, nil
}

func (s *stream) SendDatagram(b []byte) error {
	// TODO: reject if datagrams are not negotiated (yet)
	return s.datagrams.Send(b)
}

func (s *stream) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	// TODO: reject if datagrams are not negotiated (yet)
	return s.datagrams.Receive(ctx)
}

type tracingReader struct {
	io.Reader
	first *bool
	trace *httptrace.ClientTrace
}

func (r *tracingReader) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if n > 0 && r.first != nil && !*r.first {
		traceGotFirstResponseByte(r.trace)
		*r.first = true
	}
	return n, err
}
