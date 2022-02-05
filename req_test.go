package req

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func tc() *Client {
	return C().EnableDebugLog()
}

func getTestDataPath() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, ".testdata")
}

func createTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func createPostServer(t *testing.T) *httptest.Server {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Write([]byte("TestPost: text response"))
		}
	})
	return ts
}

func createGetServer(t *testing.T) *httptest.Server {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			switch r.URL.Path {
			case "/":
				w.Write([]byte("TestGet: text response"))
			case "/no-content":
				w.Write([]byte(""))
			case "/json":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"TestGet": "JSON response"}`))
			case "/json-invalid":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("TestGet: Invalid JSON"))
			case "/long-text":
				w.Write([]byte("TestGet: text response with size > 30"))
			case "/long-json":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"TestGet": "JSON response with size > 30"}`))
			case "/bad-request":
				w.WriteHeader(http.StatusBadRequest)
			case "/host-header":
				w.Write([]byte(r.Host))
			}
		}
	})

	return ts
}

func assertStatus(t *testing.T, resp *Response, err error, statusCode int, status string) {
	assertError(t, err)
	assertNotNil(t, resp)
	assertNotNil(t, resp.Body)
	assertEqual(t, statusCode, resp.StatusCode)
	assertEqual(t, status, resp.Status)
}

func assertResponse(t *testing.T, resp *Response, err error) {
	assertError(t, err)
	assertNotNil(t, resp)
	assertNotNil(t, resp.Body)
	assertEqual(t, http.StatusOK, resp.StatusCode)
	assertEqual(t, "200 OK", resp.Status)
	assertEqual(t, "HTTP/1.1", resp.Proto)
}

func assertNil(t *testing.T, v interface{}) {
	if !isNil(v) {
		t.Errorf("[%v] was expected to be nil", v)
	}
}

func assertNotNil(t *testing.T, v interface{}) {
	if isNil(v) {
		t.Errorf("[%v] was expected to be non-nil", v)
	}
}

func assertType(t *testing.T, typ, v interface{}) {
	if reflect.DeepEqual(reflect.TypeOf(typ), reflect.TypeOf(v)) {
		t.Errorf("Expected type %t, got %t", typ, v)
	}
}

func assertError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func assertEqual(t *testing.T, e, g interface{}) (r bool) {
	if !equal(e, g) {
		t.Errorf("Expected [%v], got [%v]", e, g)
	}

	return
}

func assertNotEqual(t *testing.T, e, g interface{}) (r bool) {
	if equal(e, g) {
		t.Errorf("Expected [%v], got [%v]", e, g)
	} else {
		r = true
	}

	return
}

func equal(expected, got interface{}) bool {
	return reflect.DeepEqual(expected, got)
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && rv.IsNil() {
		return true
	}
	return false
}
