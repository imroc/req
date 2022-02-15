package req

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"go/token"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"unsafe"
)

func tc() *Client {
	return C().
		SetBaseURL(getTestServerURL()).
		EnableInsecureSkipVerify()
}

var testDataPath string

func init() {
	pwd, _ := os.Getwd()
	testDataPath = filepath.Join(pwd, ".testdata")
}

func createTestServer() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(handleHTTP))
	server.EnableHTTP2 = true
	server.StartTLS()
	return server
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	}
}

var testServerMu sync.Mutex
var testServer *httptest.Server

func getTestServerURL() string {
	if testServer != nil {
		return testServer.URL
	}
	testServerMu.Lock()
	defer testServerMu.Unlock()
	testServer = createTestServer()
	return testServer.URL
}

func getTestFileContent(t *testing.T, filename string) []byte {
	b, err := ioutil.ReadFile(getTestFilePath(filename))
	assertError(t, err)
	return b
}

func getTestFilePath(filename string) string {
	return filepath.Join(testDataPath, filename)
}

type echo struct {
	Header http.Header `json:"header" xml:"header"`
	Body   string      `json:"body" xml:"body"`
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Write([]byte("TestPost: text response"))
	case "/raw-upload":
		io.Copy(ioutil.Discard, r.Body)
	case "/form":
		r.ParseForm()
		ret, _ := json.Marshal(&r.Form)
		w.Header().Set(hdrContentTypeKey, jsonContentType)
		w.Write(ret)
	case "/multipart":
		r.ParseMultipartForm(10e6)
		m := make(map[string]interface{})
		m["values"] = r.MultipartForm.Value
		m["files"] = r.MultipartForm.File
		ret, _ := json.Marshal(&m)
		w.Header().Set(hdrContentTypeKey, jsonContentType)
		w.Write(ret)
	case "/search":
		handleSearch(w, r)
	case "/redirect":
		w.Header().Set(hdrLocationKey, "/")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/echo":
		b, _ := ioutil.ReadAll(r.Body)
		e := echo{
			Header: r.Header,
			Body:   string(b),
		}
		w.Header().Set(hdrContentTypeKey, jsonContentType)
		result, _ := json.Marshal(&e)
		w.Write(result)
	}
}

func handleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	user := strings.TrimLeft(r.URL.Path, "/user")
	user = strings.TrimSuffix(user, "/profile")
	w.Write([]byte(fmt.Sprintf("%s's profile", user)))
}

type UserInfo struct {
	Username string `json:"username" xml:"username"`
	Email    string `json:"email" xml:"email"`
}

type ErrorMessage struct {
	ErrorCode    int    `json:"error_code" xml:"ErrorCode"`
	ErrorMessage string `json:"error_message" xml:"ErrorMessage"`
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	tp := r.FormValue("type")
	var marshalFunc func(v interface{}) ([]byte, error)
	if tp == "xml" {
		w.Header().Set(hdrContentTypeKey, xmlContentType)
		marshalFunc = xml.Marshal
	} else {
		w.Header().Set(hdrContentTypeKey, jsonContentType)
		marshalFunc = json.Marshal
	}
	var result interface{}
	switch username {
	case "":
		w.WriteHeader(http.StatusBadRequest)
		result = &ErrorMessage{
			ErrorCode:    10000,
			ErrorMessage: "need username",
		}
	case "imroc":
		w.WriteHeader(http.StatusOK)
		result = &UserInfo{
			Username: "imroc",
			Email:    "roc@imroc.cc",
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		result = &ErrorMessage{
			ErrorCode:    10001,
			ErrorMessage: "username not exists",
		}
	}
	data, _ := marshalFunc(result)
	w.Write(data)
}

func toGbk(s string) []byte {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		panic(e)
	}
	return d
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Write([]byte("TestGet: text response"))
	case "/bad-request":
		w.WriteHeader(http.StatusBadRequest)
	case "/host-header":
		w.Write([]byte(r.Host))
	case "/unlimited-redirect":
		w.Header().Set("Location", "/unlimited-redirect")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/redirect-to-other":
		w.Header().Set("Location", "http://dummy.local/test")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/pragma":
		w.Header().Add("Pragma", "no-cache")
	case "/gbk":
		w.Header().Set(hdrContentTypeKey, "text/plain; charset=gbk")
		w.Write(toGbk("我是roc"))
	case "/gbk-no-charset":
		b, err := ioutil.ReadFile(getTestFilePath("sample-gbk.html"))
		if err != nil {
			panic(err)
		}
		w.Header().Set(hdrContentTypeKey, "text/html")
		w.Write(b)
	case "/header":
		b, _ := json.Marshal(r.Header)
		w.Header().Set(hdrContentTypeKey, jsonContentType)
		w.Write(b)
	case "/user-agent":
		w.Write([]byte(r.Header.Get(hdrUserAgentKey)))
	case "/content-type":
		w.Write([]byte(r.Header.Get(hdrContentTypeKey)))
	case "/query-parameter":
		w.Write([]byte(r.URL.RawQuery))
	case "/search":
		handleSearch(w, r)
	default:
		if strings.HasPrefix(r.URL.Path, "/user") {
			handleGetUserProfile(w, r)
		}
	}
}

func assertStatus(t *testing.T, resp *Response, err error, statusCode int, status string) {
	assertError(t, err)
	assertNotNil(t, resp)
	assertNotNil(t, resp.Body)
	assertEqual(t, statusCode, resp.StatusCode)
	assertEqual(t, status, resp.Status)
}

func assertSuccess(t *testing.T, resp *Response, err error) {
	assertError(t, err)
	assertNotNil(t, resp)
	assertNotNil(t, resp.Body)
	assertEqual(t, http.StatusOK, resp.StatusCode)
	assertEqual(t, "200 OK", resp.Status)
	if !resp.IsSuccess() {
		t.Error("Response.IsSuccess should return true")
	}
}

func assertIsError(t *testing.T, resp *Response, err error) {
	assertError(t, err)
	assertNotNil(t, resp)
	assertNotNil(t, resp.Body)
	if !resp.IsError() {
		t.Error("Response.IsError should return true")
	}
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

func assertContains(t *testing.T, s, substr string, shouldContain bool) {
	s = strings.ToLower(s)
	isContain := strings.Contains(s, substr)
	if shouldContain {
		if !isContain {
			t.Errorf("%q is not included in %s", substr, s)
		}
	} else {
		if isContain {
			t.Errorf("%q is included in %s", substr, s)
		}
	}
}

func assertError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func assertEqualStruct(t *testing.T, e, g interface{}, onlyExported bool, excludes ...string) {
	ev := reflect.ValueOf(e).Elem()
	gv := reflect.ValueOf(g).Elem()
	et := ev.Type()
	gt := gv.Type()
	m := map[string]bool{}
	for _, exclude := range excludes {
		m[exclude] = true
	}
	if et.Kind() != reflect.Struct {
		t.Fatalf("expect object should be struct instead of %v", et.Kind().String())
	}

	if gt.Kind() != reflect.Struct {
		t.Fatalf("got object should be struct instead of %v", gt.Kind().String())
	}

	if et.Name() != gt.Name() {
		t.Fatalf("Expected type [%s], got [%s]", et.Name(), gt.Name())
	}

	for i := 0; i < ev.NumField(); i++ {
		sf := ev.Field(i)
		if sf.Kind() == reflect.Func || sf.Kind() == reflect.Slice {
			continue
		}
		st := et.Field(i)
		if m[st.Name] {
			continue
		}
		if onlyExported && !token.IsExported(st.Name) {
			continue
		}
		var ee, gg interface{}
		if !token.IsExported(st.Name) {
			ee = reflect.NewAt(sf.Type(), unsafe.Pointer(sf.UnsafeAddr())).Elem().Interface()
			gg = reflect.NewAt(sf.Type(), unsafe.Pointer(gv.Field(i).UnsafeAddr())).Elem().Interface()
		} else {
			ee = sf.Interface()
			gg = gv.Field(i).Interface()
		}
		if !reflect.DeepEqual(ee, gg) {
			t.Errorf("Field %s.%s is not equal, expected [%v], got [%v]", et.Name(), et.Field(i).Name, ee, gg)
		}
	}

}

func assertEqual(t *testing.T, e, g interface{}) {
	if !equal(e, g) {
		t.Errorf("Expected [%+v], got [%+v]", e, g)
	}
	return
}

func removeEmptyString(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
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
