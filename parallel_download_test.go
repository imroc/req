package req

import (
	"bytes"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/imroc/req/v3/internal/tests"
)

func parallelDownload(t *testing.T, handler http.HandlerFunc, segmentSize int64) ([]byte, error) {
	server := httptest.NewServer(handler)
	defer server.Close()
	output := filepath.Join(t.TempDir(), "out")
	err := C().NewParallelDownload(server.URL).
		SetConcurrency(3).
		SetSegmentSize(segmentSize).
		SetTempRootDir(t.TempDir()).
		SetOutputFile(output).
		Do()
	data, _ := os.ReadFile(output)
	return data, err
}

func TestParallelDownload(t *testing.T) {
	content := make([]byte, 3000)
	rand.Read(content)
	withRange := func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "file", time.Time{}, bytes.NewReader(content))
	}
	withoutRange := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		w.Write(content)
	}

	data, err := parallelDownload(t, withRange, 1000)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, true, bytes.Equal(content, data))

	// The server answers every segment with the whole file.
	_, err = parallelDownload(t, withoutRange, 1000)
	tests.AssertNotNil(t, err)

	// A single segment can be served without Range support.
	data, err = parallelDownload(t, withoutRange, 4000)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, true, bytes.Equal(content, data))

	notFound := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "9")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}
	_, err = parallelDownload(t, notFound, 1000)
	tests.AssertNotNil(t, err)
}
