package req

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/imroc/req/v3/internal/tests"
)

func TestRateLimitedReadCloser(t *testing.T) {
	const size = 20000
	const limit = 40000 // bytes per second, so ~0.5s for the payload
	src := bytes.Repeat([]byte("x"), size)

	r := newRateLimitedReadCloser(io.NopCloser(bytes.NewReader(src)), limit)
	start := time.Now()
	got, err := io.ReadAll(r)
	elapsed := time.Since(start)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, size, len(got))
	if !bytes.Equal(src, got) {
		t.Fatal("data read through the limiter does not match the source")
	}

	min := time.Duration(float64(size)/limit*float64(time.Second)) / 2
	if elapsed < min {
		t.Errorf("expected the read to be throttled to at least %s, but it took %s", min, elapsed)
	}
}

func TestSetDownloadLimit(t *testing.T) {
	const size = 40 * 1024
	const limit = 80 * 1024 // ~0.5s for the payload
	payload := bytes.Repeat([]byte("a"), size)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer ts.Close()

	// Without a limit the download should be near instant.
	start := time.Now()
	resp, err := C().R().Get(ts.URL)
	fast := time.Since(start)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, size, len(resp.Bytes()))
	if fast > 200*time.Millisecond {
		t.Fatalf("unlimited download was unexpectedly slow (%s), test environment too slow to be meaningful", fast)
	}

	start = time.Now()
	resp, err = C().R().SetDownloadLimit(limit).Get(ts.URL)
	slow := time.Since(start)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, size, len(resp.Bytes()))
	if !bytes.Equal(payload, resp.Bytes()) {
		t.Fatal("limited download body does not match")
	}
	min := time.Duration(float64(size)/limit*float64(time.Second)) / 2
	if slow < min {
		t.Errorf("expected the download to be throttled to at least %s, but it took %s", min, slow)
	}
}

func TestSetUploadLimit(t *testing.T) {
	const size = 40 * 1024
	const limit = 80 * 1024 // ~0.5s for the payload
	var received int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, _ := io.Copy(io.Discard, r.Body)
		received = n
	}))
	defer ts.Close()

	body := strings.Repeat("b", size)
	start := time.Now()
	resp, err := C().R().SetUploadLimit(limit).SetBodyString(body).Post(ts.URL)
	elapsed := time.Since(start)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)
	tests.AssertEqual(t, int64(size), received)

	min := time.Duration(float64(size)/limit*float64(time.Second)) / 2
	if elapsed < min {
		t.Errorf("expected the upload to be throttled to at least %s, but it took %s", min, elapsed)
	}
}

func TestSetLimitDisable(t *testing.T) {
	r := C().R().SetDownloadLimit(1024).SetUploadLimit(1024)
	tests.AssertEqual(t, int64(1024), r.downloadLimit)
	tests.AssertEqual(t, int64(1024), r.uploadLimit)

	// A zero or negative value clears the limit.
	r.SetDownloadLimit(0).SetUploadLimit(-1)
	tests.AssertEqual(t, int64(0), r.downloadLimit)
	tests.AssertEqual(t, int64(0), r.uploadLimit)
}
