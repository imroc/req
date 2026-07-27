package req

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imroc/req/v3/internal/tests"
)

func TestMaxResponseSizeWithinLimit(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		// "TestGet: text response" is 23 bytes
		resp, err := c.SetMaxResponseSize(100).R().Get("/")
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, "TestGet: text response", resp.String())
	})
}

func TestMaxResponseSizeExactLimit(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		const size = 64
		resp, err := c.SetMaxResponseSize(size).R().Get("/fixed-size?size=64")
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, size, len(resp.Bytes()))
	})
}

func TestMaxResponseSizeContentLengthExceeded(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		resp, err := c.SetMaxResponseSize(10).R().Get("/fixed-size?size=100")
		if err == nil {
			t.Fatal("expected error when Content-Length exceeds limit")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		var tooLarge *ResponseBodyTooLargeError
		if !errors.As(err, &tooLarge) {
			t.Fatalf("expected *ResponseBodyTooLargeError, got %T", err)
		}
		tests.AssertEqual(t, int64(10), tooLarge.Limit)
		tests.AssertEqual(t, int64(100), tooLarge.ContentLength)
		// Response headers/status should still be available.
		tests.AssertEqual(t, http.StatusOK, resp.StatusCode)
		// Body must not have been buffered.
		tests.AssertEqual(t, 0, len(resp.Bytes()))
	})
}

func TestMaxResponseSizeChunkedExceeded(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		// Chunked response has no Content-Length; limit is enforced while reading.
		resp, err := c.SetMaxResponseSize(50).R().Get("/chunked-size?size=200")
		if err == nil {
			t.Fatal("expected error when chunked body exceeds limit")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		var tooLarge *ResponseBodyTooLargeError
		if !errors.As(err, &tooLarge) {
			t.Fatalf("expected *ResponseBodyTooLargeError, got %T", err)
		}
		tests.AssertEqual(t, int64(50), tooLarge.Limit)
		tests.AssertEqual(t, int64(-1), tooLarge.ContentLength)
		tests.AssertEqual(t, http.StatusOK, resp.StatusCode)
		// Partial body may have been read up to the limit.
		if len(resp.Bytes()) > 50 {
			t.Fatalf("buffered body larger than limit: %d", len(resp.Bytes()))
		}
	})
}

func TestMaxResponseSizeChunkedWithinLimit(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		resp, err := c.SetMaxResponseSize(500).R().Get("/chunked-size?size=100")
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, 100, len(resp.Bytes()))
	})
}

func TestMaxResponseSizeRequestOverridesClient(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		c.SetMaxResponseSize(10)
		// Request-level higher limit allows the larger body.
		resp, err := c.R().SetMaxResponseSize(200).Get("/fixed-size?size=100")
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, 100, len(resp.Bytes()))

		// Request-level lower limit rejects a body the client would accept.
		c.SetMaxResponseSize(1000)
		resp, err = c.R().SetMaxResponseSize(10).Get("/fixed-size?size=100")
		if err == nil {
			t.Fatal("expected request-level limit to reject response")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		_ = resp
	})
}

func TestMaxResponseSizeRequestDisablesClientLimit(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		c.SetMaxResponseSize(10)
		// 0 on the request disables the limit for this request.
		resp, err := c.R().SetMaxResponseSize(0).Get("/fixed-size?size=100")
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, 100, len(resp.Bytes()))
	})
}

func TestMaxResponseSizeDisableAutoRead(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		c.SetMaxResponseSize(50).DisableAutoReadResponse()
		resp, err := c.R().Get("/chunked-size?size=200")
		// Headers succeed; body is not auto-read.
		tests.AssertNoError(t, err)
		tests.AssertEqual(t, http.StatusOK, resp.StatusCode)

		// Manual read hits the limit.
		_, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			t.Fatal("expected error when reading oversized body manually")
		}
		if !errors.Is(readErr, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", readErr)
		}
		_ = resp.Body.Close()
	})
}

func TestMaxResponseSizeToBytes(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		c.SetMaxResponseSize(50).DisableAutoReadResponse()
		resp, err := c.R().Get("/chunked-size?size=200")
		tests.AssertNoError(t, err)

		_, err = resp.ToBytes()
		if err == nil {
			t.Fatal("expected ToBytes to fail when body exceeds limit")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		if !errors.Is(resp.Err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected resp.Err to be ErrResponseBodyTooLarge, got %v", resp.Err)
		}
	})
}

func TestMaxResponseSizeDownload(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		dir := t.TempDir()
		out := filepath.Join(dir, "out.bin")

		// Content-Length over limit: reject without writing a full file.
		resp, err := c.SetMaxResponseSize(10).R().
			SetOutputFile(out).
			Get("/fixed-size?size=100")
		if err == nil {
			t.Fatal("expected download to fail when Content-Length exceeds limit")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		_ = resp
		// File should not have been created (handleDownload skipped on prior error).
		if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
			t.Fatalf("expected output file not to be created, stat err=%v", statErr)
		}

		// Chunked body over limit: download fails mid-stream.
		out2 := filepath.Join(dir, "out2.bin")
		resp, err = c.SetMaxResponseSize(50).R().
			SetOutputFile(out2).
			Get("/chunked-size?size=200")
		if err == nil {
			t.Fatal("expected download to fail when body exceeds limit during read")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		_ = resp
	})
}

func TestMaxResponseSizeDownloadWriter(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		var buf bytes.Buffer
		resp, err := c.SetMaxResponseSize(50).R().
			SetOutput(&buf).
			Get("/chunked-size?size=200")
		if err == nil {
			t.Fatal("expected error when writing oversized body to output")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		if buf.Len() > 50 {
			t.Fatalf("wrote more than limit: %d", buf.Len())
		}
		_ = resp
	})
}

func TestMaxResponseSizeLargeContentLengthNoFullRead(t *testing.T) {
	// Regression for the original bandwidth concern: a huge Content-Length must
	// be rejected without buffering the body.
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		resp, err := c.SetMaxResponseSize(1024).R().Get("/download")
		if err == nil {
			t.Fatal("expected error for 100MiB Content-Length with 1KiB limit")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		var tooLarge *ResponseBodyTooLargeError
		if !errors.As(err, &tooLarge) {
			t.Fatalf("expected *ResponseBodyTooLargeError, got %T", err)
		}
		tests.AssertEqual(t, int64(1024), tooLarge.Limit)
		if tooLarge.ContentLength <= 1024 {
			t.Fatalf("expected ContentLength > limit, got %d", tooLarge.ContentLength)
		}
		tests.AssertEqual(t, 0, len(resp.Bytes()))
	})
}

func TestMaxResponseSizeZeroMeansUnlimited(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		resp, err := c.SetMaxResponseSize(0).R().Get("/fixed-size?size=100")
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, 100, len(resp.Bytes()))
	})
}

func TestMaxResponseSizeNegativeTreatedAsUnlimited(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		resp, err := c.SetMaxResponseSize(-1).R().Get("/fixed-size?size=100")
		assertSuccess(t, resp, err)
		tests.AssertEqual(t, 100, len(resp.Bytes()))
	})
}

func TestMaxResponseSizeClone(t *testing.T) {
	c := tc().SetMaxResponseSize(42)
	cc := c.Clone()
	tests.AssertEqual(t, int64(42), cc.maxResponseSize)

	// Changing the clone must not affect the original.
	cc.SetMaxResponseSize(99)
	tests.AssertEqual(t, int64(42), c.maxResponseSize)
	tests.AssertEqual(t, int64(99), cc.maxResponseSize)
}

func TestMaxResponseSizeErrorMessage(t *testing.T) {
	e1 := &ResponseBodyTooLargeError{Limit: 10, ContentLength: 100}
	if !strings.Contains(e1.Error(), "Content-Length 100") {
		t.Fatalf("unexpected error message: %s", e1.Error())
	}
	if !strings.Contains(e1.Error(), "limit 10") {
		t.Fatalf("unexpected error message: %s", e1.Error())
	}

	e2 := &ResponseBodyTooLargeError{Limit: 50, ContentLength: -1}
	if !strings.Contains(e2.Error(), "exceeds limit of 50") {
		t.Fatalf("unexpected error message: %s", e2.Error())
	}
	if !errors.Is(e1, ErrResponseBodyTooLarge) || !errors.Is(e2, ErrResponseBodyTooLarge) {
		t.Fatal("errors.Is should match ErrResponseBodyTooLarge")
	}
}

func TestMaxResponseBodyReaderStickyError(t *testing.T) {
	r := &maxResponseBodyReader{
		r:     io.NopCloser(bytes.NewReader(bytes.Repeat([]byte{'a'}, 100))),
		n:     10,
		limit: 10,
	}
	buf := make([]byte, 64)
	n, err := r.Read(buf)
	if err == nil {
		t.Fatal("expected error on first oversized read")
	}
	if n > 10 {
		t.Fatalf("read more than limit: %d", n)
	}
	if !errors.Is(err, ErrResponseBodyTooLarge) {
		t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
	}
	// Sticky: subsequent reads return the same error.
	_, err2 := r.Read(buf)
	if !errors.Is(err2, ErrResponseBodyTooLarge) {
		t.Fatalf("expected sticky ErrResponseBodyTooLarge, got %v", err2)
	}
}

func TestMaxResponseBodyReaderExactLimit(t *testing.T) {
	data := bytes.Repeat([]byte{'b'}, 10)
	r := &maxResponseBodyReader{
		r:     io.NopCloser(bytes.NewReader(data)),
		n:     10,
		limit: 10,
	}
	got, err := io.ReadAll(r)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, data, got)
}

func TestMaxResponseSizeWithSuccessResult(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		var result map[string]string
		resp, err := c.SetMaxResponseSize(5).R().
			SetSuccessResult(&result).
			Get("/json")
		if err == nil {
			t.Fatal("expected error when JSON body exceeds limit")
		}
		if !errors.Is(err, ErrResponseBodyTooLarge) {
			t.Fatalf("expected ErrResponseBodyTooLarge, got %v", err)
		}
		_ = resp
	})
}
