package req

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/tests"
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
	"strconv"
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
	w.Header().Add("Method", r.Method)
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
	b, err := ioutil.ReadFile(tests.GetTestFilePath(filename))
	tests.AssertNoError(t, err)
	return b
}

func assertClone(t *testing.T, e, g interface{}) {
	ev := reflect.ValueOf(e).Elem()
	gv := reflect.ValueOf(g).Elem()
	et := ev.Type()

	for i := 0; i < ev.NumField(); i++ {
		sf := ev.Field(i)
		st := et.Field(i)

		var ee, gg interface{}
		if !token.IsExported(st.Name) {
			ee = reflect.NewAt(sf.Type(), unsafe.Pointer(sf.UnsafeAddr())).Elem().Interface()
			gg = reflect.NewAt(sf.Type(), unsafe.Pointer(gv.Field(i).UnsafeAddr())).Elem().Interface()
		} else {
			ee = sf.Interface()
			gg = gv.Field(i).Interface()
		}
		if sf.Kind() == reflect.Func || sf.Kind() == reflect.Slice || sf.Kind() == reflect.Ptr {
			if ee != nil {
				if gg == nil {
					t.Errorf("Field %s.%s is nil", et.Name(), et.Field(i).Name)
				}
			}
			continue
		}
		if !reflect.DeepEqual(ee, gg) {
			t.Errorf("Field %s.%s is not equal, expected [%v], got [%v]", et.Name(), et.Field(i).Name, ee, gg)
		}
	}
}

// Echo is used in "/echo" API.
type Echo struct {
	Header http.Header `json:"header" xml:"header"`
	Body   string      `json:"body" xml:"body"`
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		io.Copy(ioutil.Discard, r.Body)
		w.Write([]byte("TestPost: text response"))
	case "/raw-upload":
		io.Copy(ioutil.Discard, r.Body)
	case "/file-text":
		r.ParseMultipartForm(10e6)
		files := r.MultipartForm.File["file"]
		file, _ := files[0].Open()
		b, _ := ioutil.ReadAll(file)
		r.ParseForm()
		if a := r.FormValue("attempt"); a != "" && a != "2" {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(b)
	case "/form":
		r.ParseForm()
		ret, _ := json.Marshal(&r.Form)
		w.Header().Set(header.ContentType, header.JsonContentType)
		w.Write(ret)
	case "/multipart":
		r.ParseMultipartForm(10e6)
		m := make(map[string]interface{})
		m["values"] = r.MultipartForm.Value
		m["files"] = r.MultipartForm.File
		ret, _ := json.Marshal(&m)
		w.Header().Set(header.ContentType, header.JsonContentType)
		w.Write(ret)
	case "/search":
		handleSearch(w, r)
	case "/redirect":
		io.Copy(ioutil.Discard, r.Body)
		w.Header().Set(header.Location, "/")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/content-type":
		io.Copy(ioutil.Discard, r.Body)
		w.Write([]byte(r.Header.Get(header.ContentType)))
	case "/echo":
		b, _ := ioutil.ReadAll(r.Body)
		e := Echo{
			Header: r.Header,
			Body:   string(b),
		}
		w.Header().Set(header.ContentType, header.JsonContentType)
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
		w.Header().Set(header.ContentType, header.XmlContentType)
		marshalFunc = xml.Marshal
	} else {
		w.Header().Set(header.ContentType, header.JsonContentType)
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
	case "/too-many":
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set(header.ContentType, header.JsonContentType)
		w.Write([]byte(`{"errMsg":"too many requests"}`))
	case "/chunked":
		w.Header().Add("Trailer", "Expires")
		w.Write([]byte(`This is a chunked body`))
	case "/host-header":
		w.Write([]byte(r.Host))
	case "/json":
		r.ParseForm()
		if r.FormValue("type") != "no" {
			w.Header().Set(header.ContentType, header.JsonContentType)
		}
		w.Header().Set(header.ContentType, header.JsonContentType)
		if r.FormValue("error") == "yes" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "not allowed"}`))
		} else {
			w.Write([]byte(`{"name": "roc"}`))
		}
	case "/xml":
		r.ParseForm()
		if r.FormValue("type") != "no" {
			w.Header().Set(header.ContentType, header.XmlContentType)
		}
		w.Write([]byte(`<user><name>roc</name></user>`))
	case "/unlimited-redirect":
		w.Header().Set("Location", "/unlimited-redirect")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/redirect-to-other":
		w.Header().Set("Location", "http://dummy.local/test")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/pragma":
		w.Header().Add("Pragma", "no-cache")
	case "/payload":
		b, _ := ioutil.ReadAll(r.Body)
		w.Write(b)
	case "/gbk":
		w.Header().Set(header.ContentType, "text/plain; charset=gbk")
		w.Write(toGbk("我是roc"))
	case "/gbk-no-charset":
		b, err := ioutil.ReadFile(tests.GetTestFilePath("sample-gbk.html"))
		if err != nil {
			panic(err)
		}
		w.Header().Set(header.ContentType, "text/html")
		w.Write(b)
	case "/header":
		b, _ := json.Marshal(r.Header)
		w.Header().Set(header.ContentType, header.JsonContentType)
		w.Write(b)
	case "/user-agent":
		w.Write([]byte(r.Header.Get(header.UserAgent)))
	case "/content-type":
		w.Write([]byte(r.Header.Get(header.ContentType)))
	case "/query-parameter":
		w.Write([]byte(r.URL.RawQuery))
	case "/search":
		handleSearch(w, r)
	case "/download":
		size := 100 * 1024 * 1024
		w.Header().Set("Content-Length", strconv.Itoa(size))
		buf := make([]byte, 1024)
		for i := 0; i < 1024; i++ {
			buf[i] = 'h'
		}
		for i := 0; i < size; {
			wbuf := buf
			if size-i < 1024 {
				wbuf = buf[:size-i]
			}
			n, err := w.Write(wbuf)
			if err != nil {
				break
			}
			i += n
		}
	case "/protected":
		auth := r.Header.Get("Authorization")
		if auth == "Bearer goodtoken" {
			w.Write([]byte("good"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`bad`))
		}
	default:
		if strings.HasPrefix(r.URL.Path, "/user") {
			handleGetUserProfile(w, r)
		}
	}
}

func assertStatus(t *testing.T, resp *Response, err error, statusCode int, status string) {
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, resp)
	tests.AssertNotNil(t, resp.Body)
	tests.AssertEqual(t, statusCode, resp.StatusCode)
	tests.AssertEqual(t, status, resp.Status)
}

func assertSuccess(t *testing.T, resp *Response, err error) {
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, resp.Response)
	tests.AssertNotNil(t, resp.Response.Body)
	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)
	tests.AssertEqual(t, "200 OK", resp.Status)
	if !resp.IsSuccess() {
		t.Error("Response.IsSuccess should return true")
	}
}

func assertIsError(t *testing.T, resp *Response, err error) {
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, resp)
	tests.AssertNotNil(t, resp.Body)
	if !resp.IsError() {
		t.Error("Response.IsError should return true")
	}
}

func TestTrailer(t *testing.T) {
	resp, err := tc().EnableForceHTTP1().R().Get("/chunked")
	assertSuccess(t, resp, err)
	_, ok := resp.Trailer["Expires"]
	if !ok {
		t.Error("trailer not exists")
	}
}

func testWithAllTransport(t *testing.T, testFunc func(t *testing.T, c *Client)) {
	testFunc(t, tc())
	testFunc(t, tc().EnableForceHTTP1())
}
