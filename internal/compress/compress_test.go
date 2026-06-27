package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
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
