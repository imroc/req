package req

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"go/token"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
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
	case "/file-text":
		r.ParseMultipartForm(10e6)
		files := r.MultipartForm.File["file"]
		file, _ := files[0].Open()
		b, _ := ioutil.ReadAll(file)
		w.Write(b)
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
	case "/payload":
		b, _ := ioutil.ReadAll(r.Body)
		w.Write(b)
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

func testGlobalWrapperEnableDumps(t *testing.T) {
	testGlobalWrapperEnableDump(t, func(reqHeader, reqBody, respHeader, respBody *bool) *Request {
		*reqHeader = true
		*reqBody = true
		*respHeader = true
		*respBody = true
		return EnableDump()
	})

	testGlobalWrapperEnableDump(t, func(reqHeader, reqBody, respHeader, respBody *bool) *Request {
		*reqHeader = false
		*reqBody = false
		*respHeader = true
		*respBody = true
		return EnableDumpWithoutRequest()
	})

	testGlobalWrapperEnableDump(t, func(reqHeader, reqBody, respHeader, respBody *bool) *Request {
		*reqHeader = true
		*reqBody = false
		*respHeader = true
		*respBody = true
		return EnableDumpWithoutRequestBody()
	})

	testGlobalWrapperEnableDump(t, func(reqHeader, reqBody, respHeader, respBody *bool) *Request {
		*reqHeader = true
		*reqBody = true
		*respHeader = false
		*respBody = false
		return EnableDumpWithoutResponse()
	})

	testGlobalWrapperEnableDump(t, func(reqHeader, reqBody, respHeader, respBody *bool) *Request {
		*reqHeader = true
		*reqBody = true
		*respHeader = true
		*respBody = false
		return EnableDumpWithoutResponseBody()
	})

	testGlobalWrapperEnableDump(t, func(reqHeader, reqBody, respHeader, respBody *bool) *Request {
		*reqHeader = false
		*reqBody = true
		*respHeader = false
		*respBody = true
		return EnableDumpWithoutHeader()
	})

	testGlobalWrapperEnableDump(t, func(reqHeader, reqBody, respHeader, respBody *bool) *Request {
		*reqHeader = true
		*reqBody = false
		*respHeader = true
		*respBody = false
		return EnableDumpWithoutBody()
	})
}

func testGlobalWrapperEnableDump(t *testing.T, fn func(reqHeader, reqBody, respHeader, respBody *bool) *Request) {
	var reqHeader, reqBody, respHeader, respBody bool
	r := fn(&reqHeader, &reqBody, &respHeader, &respBody)
	resp, err := r.SetBody(`test body`).Post(getTestServerURL() + "/")
	assertSuccess(t, resp, err)
	dump := resp.Dump()
	assertContains(t, dump, "user-agent", reqHeader)
	assertContains(t, dump, "test body", reqBody)
	assertContains(t, dump, "date", respHeader)
	assertContains(t, dump, "testpost: text response", respBody)
}

func testGlobalWrapperSendRequest(t *testing.T) {
	testURL := getTestServerURL() + "/"

	resp, err := Put(testURL)
	assertSuccess(t, resp, err)
	assertEqual(t, "PUT", resp.Header.Get("Method"))
	resp = MustPut(testURL)
	assertEqual(t, "PUT", resp.Header.Get("Method"))

	resp, err = Patch(testURL)
	assertSuccess(t, resp, err)
	assertEqual(t, "PATCH", resp.Header.Get("Method"))
	resp = MustPatch(testURL)
	assertEqual(t, "PATCH", resp.Header.Get("Method"))

	resp, err = Delete(testURL)
	assertSuccess(t, resp, err)
	assertEqual(t, "DELETE", resp.Header.Get("Method"))
	resp = MustDelete(testURL)
	assertEqual(t, "DELETE", resp.Header.Get("Method"))

	resp, err = Options(testURL)
	assertSuccess(t, resp, err)
	assertEqual(t, "OPTIONS", resp.Header.Get("Method"))
	resp = MustOptions(testURL)
	assertEqual(t, "OPTIONS", resp.Header.Get("Method"))

	resp, err = Head(testURL)
	assertSuccess(t, resp, err)
	assertEqual(t, "HEAD", resp.Header.Get("Method"))
	resp = MustHead(testURL)
	assertEqual(t, "HEAD", resp.Header.Get("Method"))

	resp, err = Get(testURL)
	assertSuccess(t, resp, err)
	assertEqual(t, "GET", resp.Header.Get("Method"))
	resp = MustGet(testURL)
	assertEqual(t, "GET", resp.Header.Get("Method"))

	resp, err = Post(testURL)
	assertSuccess(t, resp, err)
	assertEqual(t, "POST", resp.Header.Get("Method"))
	resp = MustPost(testURL)
	assertEqual(t, "POST", resp.Header.Get("Method"))
}

func TestGlobalWrapper(t *testing.T) {
	EnableInsecureSkipVerify()
	testGlobalWrapperSendRequest(t)
	testGlobalWrapperEnableDumps(t)
	DisableInsecureSkipVerify()

	SetCookieJar(nil)
	assertEqual(t, nil, DefaultClient().httpClient.Jar)

	testErr := errors.New("test")
	testDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return nil, testErr
	}
	testDialTLS := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return nil, testErr
	}
	SetDialTLS(testDialTLS)
	SetDial(testDial)
	_, err := DefaultClient().t.DialTLSContext(nil, "", "")
	assertEqual(t, testErr, err)
	_, err = DefaultClient().t.DialContext(nil, "", "")
	assertEqual(t, testErr, err)

	timeout := 2 * time.Second
	SetTLSHandshakeTimeout(timeout)
	assertEqual(t, timeout, DefaultClient().t.TLSHandshakeTimeout)

	EnableAllowGetMethodPayload()
	assertEqual(t, true, DefaultClient().AllowGetMethodPayload)

	marshalFunc := func(v interface{}) ([]byte, error) {
		return nil, testErr
	}
	unmarshalFunc := func(data []byte, v interface{}) error {
		return testErr
	}
	SetJsonMarshal(marshalFunc)
	SetJsonUnmarshal(unmarshalFunc)
	SetXmlMarshal(marshalFunc)
	SetXmlUnmarshal(unmarshalFunc)
	_, err = DefaultClient().jsonMarshal(nil)
	assertEqual(t, testErr, err)
	err = DefaultClient().jsonUnmarshal(nil, nil)
	assertEqual(t, testErr, err)
	_, err = DefaultClient().xmlMarshal(nil)
	assertEqual(t, testErr, err)
	err = DefaultClient().xmlUnmarshal(nil, nil)
	assertEqual(t, testErr, err)

	EnableTraceAll()
	assertEqual(t, true, DefaultClient().trace)
	DisableTraceAll()
	assertEqual(t, false, DefaultClient().trace)

	len1 := len(DefaultClient().afterResponse)
	OnAfterResponse(func(client *Client, response *Response) error {
		return nil
	})
	len2 := len(DefaultClient().afterResponse)
	assertEqual(t, true, len1+1 == len2)

	OnBeforeRequest(func(client *Client, request *Request) error {
		return nil
	})
	assertEqual(t, true, len(DefaultClient().udBeforeRequest) == 1)

	SetProxyURL("http://dummy.proxy.local")
	u, err := DefaultClient().t.Proxy(nil)
	assertError(t, err)
	assertEqual(t, "http://dummy.proxy.local", u.String())

	u, _ = url.Parse("http://dummy.proxy.local")
	proxy := http.ProxyURL(u)
	SetProxy(proxy)
	uu, err := DefaultClient().t.Proxy(nil)
	assertError(t, err)
	assertEqual(t, u.String(), uu.String())

	SetCommonContentType(jsonContentType)
	assertEqual(t, jsonContentType, DefaultClient().Headers.Get(hdrContentTypeKey))

	SetCommonHeader("my-header", "my-value")
	assertEqual(t, "my-value", DefaultClient().Headers.Get("my-header"))

	SetCommonHeaders(map[string]string{
		"header1": "value1",
		"header2": "value2",
	})
	assertEqual(t, "value1", DefaultClient().Headers.Get("header1"))
	assertEqual(t, "value2", DefaultClient().Headers.Get("header2"))

	SetCommonBasicAuth("imroc", "123456")
	assertEqual(t, "Basic aW1yb2M6MTIzNDU2", DefaultClient().Headers.Get("Authorization"))

	SetCommonBearerAuthToken("123456")
	assertEqual(t, "Bearer 123456", DefaultClient().Headers.Get("Authorization"))

	SetUserAgent("test")
	assertEqual(t, "test", DefaultClient().Headers.Get(hdrUserAgentKey))

	SetTimeout(timeout)
	assertEqual(t, timeout, DefaultClient().httpClient.Timeout)

	l := createDefaultLogger()
	SetLogger(l)
	assertEqual(t, l, DefaultClient().log)

	SetScheme("https")
	assertEqual(t, "https", DefaultClient().scheme)

	EnableDebugLog()
	assertEqual(t, true, DefaultClient().DebugLog)

	DisableDebugLog()
	assertEqual(t, false, DefaultClient().DebugLog)

	SetCommonCookies(&http.Cookie{Name: "test", Value: "test"})
	assertEqual(t, "test", DefaultClient().Cookies[0].Name)

	SetCommonQueryString("test1=test1")
	assertEqual(t, "test1", DefaultClient().QueryParams.Get("test1"))

	SetCommonPathParams(map[string]string{"test1": "test1"})
	assertEqual(t, "test1", DefaultClient().PathParams["test1"])

	SetCommonPathParam("test2", "test2")
	assertEqual(t, "test2", DefaultClient().PathParams["test2"])

	AddCommonQueryParam("test1", "test11")
	assertEqual(t, []string{"test1", "test11"}, DefaultClient().QueryParams["test1"])

	SetCommonQueryParam("test1", "test111")
	assertEqual(t, "test111", DefaultClient().QueryParams.Get("test1"))

	SetCommonQueryParams(map[string]string{"test1": "test1"})
	assertEqual(t, "test1", DefaultClient().QueryParams.Get("test1"))

	EnableInsecureSkipVerify()
	assertEqual(t, true, DefaultClient().t.TLSClientConfig.InsecureSkipVerify)

	DisableInsecureSkipVerify()
	assertEqual(t, false, DefaultClient().t.TLSClientConfig.InsecureSkipVerify)

	DisableCompression()
	assertEqual(t, true, DefaultClient().t.DisableCompression)

	EnableCompression()
	assertEqual(t, false, DefaultClient().t.DisableCompression)

	DisableKeepAlives()
	assertEqual(t, true, DefaultClient().t.DisableKeepAlives)

	EnableKeepAlives()
	assertEqual(t, false, DefaultClient().t.DisableKeepAlives)

	config := GetTLSClientConfig()
	assertEqual(t, config, DefaultClient().t.TLSClientConfig)

	SetRootCertsFromFile(getTestFilePath("sample-root.pem"))
	assertEqual(t, true, DefaultClient().t.TLSClientConfig.RootCAs != nil)

	SetRootCertFromString(string(getTestFileContent(t, "sample-root.pem")))
	assertEqual(t, true, DefaultClient().t.TLSClientConfig.RootCAs != nil)

	SetCerts(tls.Certificate{}, tls.Certificate{})
	assertEqual(t, true, len(DefaultClient().t.TLSClientConfig.Certificates) == 2)

	SetCertFromFile(
		getTestFilePath("sample-client.pem"),
		getTestFilePath("sample-client-key.pem"),
	)
	assertEqual(t, true, len(DefaultClient().t.TLSClientConfig.Certificates) == 3)

	SetOutputDirectory(testDataPath)
	assertEqual(t, testDataPath, DefaultClient().outputDirectory)

	baseURL := "http://dummy-req.local/test"
	SetBaseURL(baseURL)
	assertEqual(t, baseURL, DefaultClient().BaseURL)

	form := make(url.Values)
	form.Add("test", "test")
	SetCommonFormDataFromValues(form)
	assertEqual(t, form, DefaultClient().FormData)

	SetCommonFormData(map[string]string{"test2": "test2"})
	assertEqual(t, "test2", DefaultClient().FormData.Get("test2"))

	DisableAutoReadResponse()
	assertEqual(t, true, DefaultClient().disableAutoReadResponse)
	EnableAutoReadResponse()
	assertEqual(t, false, DefaultClient().disableAutoReadResponse)

	EnableDumpAll()
	opt := DefaultClient().getDumpOptions()
	assertEqual(t, true, opt.RequestHeader == true && opt.RequestBody == true && opt.ResponseHeader == true && opt.ResponseBody == true)
	EnableDumpAllAsync()
	assertEqual(t, true, opt.Async)
	EnableDumpAllWithoutBody()
	assertEqual(t, true, opt.ResponseBody == false && opt.RequestBody == false)
	opt.ResponseBody = true
	opt.RequestBody = true
	EnableDumpAllWithoutResponse()
	assertEqual(t, true, opt.ResponseBody == false && opt.ResponseHeader == false)
	opt.ResponseBody = true
	opt.ResponseHeader = true
	EnableDumpAllWithoutRequest()
	assertEqual(t, true, opt.RequestHeader == false && opt.RequestBody == false)
	opt.RequestHeader = true
	opt.RequestBody = true
	EnableDumpAllWithoutHeader()
	assertEqual(t, true, opt.RequestHeader == false && opt.ResponseHeader == false)
	SetCommonDumpOptions(&DumpOptions{
		RequestHeader: true,
	})
	opt = DefaultClient().getDumpOptions()
	assertEqual(t, true, opt.RequestHeader == true && opt.ResponseHeader == false)
}
