package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"strings"
	"testing"
)

func gzipData(t *testing.T, data string) io.ReadCloser {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	gw.Close()
	return io.NopCloser(&buf)
}

func TestGzipReader(t *testing.T) {
	original := "hello world, this is a test string for gzip"
	body := gzipData(t, original)
	gz := NewGzipReader(body)
	defer gz.Close()

	out, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(out) != original {
		t.Fatalf("got %q, want %q", string(out), original)
	}
}

func TestGzipReaderInvalidData(t *testing.T) {
	body := io.NopCloser(strings.NewReader("not gzip data"))
	gz := NewGzipReader(body)
	defer gz.Close()

	_, err := gz.Read(make([]byte, 10))
	if err == nil {
		t.Fatal("expected error for invalid gzip data")
	}
}

func TestGzipReaderCloseTwice(t *testing.T) {
	body := gzipData(t, "test")
	gz := NewGzipReader(body)
	if err := gz.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	// After close, Read should return sticky error
	_, err := gz.Read(make([]byte, 10))
	if err == nil {
		t.Fatal("expected error after close")
	}
}

func TestGzipReaderGetSetUnderlyingBody(t *testing.T) {
	body1 := gzipData(t, "test1")
	body2 := gzipData(t, "test2")
	gz := NewGzipReader(body1)
	if gz.GetUnderlyingBody() != body1 {
		t.Fatal("GetUnderlyingBody mismatch")
	}
	gz.SetUnderlyingBody(body2)
	if gz.GetUnderlyingBody() != body2 {
		t.Fatal("SetUnderlyingBody failed")
	}
}

func deflateData(t *testing.T, data string) io.ReadCloser {
	t.Helper()
	var buf bytes.Buffer
	fw, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte(data))
	fw.Close()
	return io.NopCloser(&buf)
}

func TestDeflateReader(t *testing.T) {
	original := "hello deflate world"
	body := deflateData(t, original)
	dr := NewDeflateReader(body)
	defer dr.Close()

	out, err := io.ReadAll(dr)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(out) != original {
		t.Fatalf("got %q, want %q", string(out), original)
	}
}

func TestNewCompressReader(t *testing.T) {
	tests := []struct {
		encoding string
		notNil   bool
	}{
		{"gzip", true},
		{"deflate", true},
		{"br", true},
		{"zstd", true},
		{"unknown", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.encoding, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(""))
			cr := NewCompressReader(body, tt.encoding)
			if tt.notNil && cr == nil {
				t.Fatal("expected non-nil CompressReader")
			}
			if !tt.notNil && cr != nil {
				t.Fatal("expected nil CompressReader")
			}
		})
	}
}

func zlibData(t *testing.T, data string) io.ReadCloser {
	t.Helper()
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	zw.Close()
	return io.NopCloser(&buf)
}

// Content-Encoding: deflate is the zlib format of RFC 1950. Raw RFC 1951
// deflate is the common variant. Both have to read.
func TestDeflateReaderReadsBothVariants(t *testing.T) {
	const original = "hello deflate world, long enough to compress"
	tests := []struct {
		name string
		body func(*testing.T, string) io.ReadCloser
	}{
		{"zlib", zlibData},
		{"raw", deflateData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDeflateReader(tt.body(t, original))
			defer dr.Close()

			out, err := io.ReadAll(dr)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}
			if string(out) != original {
				t.Fatalf("got %q, want %q", string(out), original)
			}
		})
	}
}

// The header check is compared against compress/zlib itself over every
// possible two-byte header, so it cannot drift from the reference and no
// header has to be worked out by hand. A zlib header naming a preset
// dictionary is still a zlib header; the reader surfaces zlib's own error
// for it instead of reading the body as raw deflate.
func TestLooksLikeZlibAgreesWithCompressZlib(t *testing.T) {
	var accepted int
	for cmf := 0; cmf < 256; cmf++ {
		for flg := 0; flg < 256; flg++ {
			header := []byte{byte(cmf), byte(flg)}
			_, err := zlib.NewReader(bytes.NewReader(append(header, make([]byte, 8)...)))
			want := err == nil || err == zlib.ErrDictionary
			if want {
				accepted++
			}
			if got := looksLikeZlib(byte(cmf), byte(flg)); got != want {
				t.Errorf("header %#02x %#02x: looksLikeZlib = %v, compress/zlib says %v", cmf, flg, got, want)
			}
		}
	}
	if accepted == 0 {
		t.Fatal("compress/zlib accepted no header at all, the check proves nothing")
	}
}

func TestDeflateReaderCloseClosesTheBody(t *testing.T) {
	body := deflateData(t, "hello deflate world")
	tracked := &closeTracker{ReadCloser: body}
	dr := NewDeflateReader(tracked)

	if _, err := io.ReadAll(dr); err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if err := dr.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if !tracked.closed {
		t.Fatal("the response body was never closed, so the connection is never released")
	}
}

type closeTracker struct {
	io.ReadCloser
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return c.ReadCloser.Close()
}
