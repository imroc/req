package req

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func getTestDataPath() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, ".testdata")
}

func createTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func createGetServer(t *testing.T) *httptest.Server {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		if r.Method == http.MethodGet {
			switch r.URL.Path {
			case "/":
				_, _ = w.Write([]byte("TestGet: text response"))
			case "/no-content":
				_, _ = w.Write([]byte(""))
			case "/json":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"TestGet": "JSON response"}`))
			case "/json-invalid":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("TestGet: Invalid JSON"))
			case "/long-text":
				_, _ = w.Write([]byte("TestGet: text response with size > 30"))
			case "/long-json":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"TestGet": "JSON response with size > 30"}`))
			case "/mypage":
				w.WriteHeader(http.StatusBadRequest)
			case "/mypage2":
				_, _ = w.Write([]byte("TestGet: text response from mypage2"))
			case "/host-header":
				_, _ = w.Write([]byte(r.Host))
			}
		}
	})

	return ts
}
