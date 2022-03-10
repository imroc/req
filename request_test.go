package req

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/imroc/req/v3/internal/tests"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestMustSendMethods(t *testing.T) {
	c := tc()
	testCases := []struct {
		SendReq      func(req *Request, url string) *Response
		ExpectMethod string
	}{
		{
			SendReq: func(req *Request, url string) *Response {
				return req.MustGet(url)
			},
			ExpectMethod: "GET",
		},
		{
			SendReq: func(req *Request, url string) *Response {
				return req.MustPost(url)
			},
			ExpectMethod: "POST",
		},
		{
			SendReq: func(req *Request, url string) *Response {
				return req.MustPatch(url)
			},
			ExpectMethod: "PATCH",
		},
		{
			SendReq: func(req *Request, url string) *Response {
				return req.MustDelete(url)
			},
			ExpectMethod: "DELETE",
		},
		{
			SendReq: func(req *Request, url string) *Response {
				return req.MustOptions(url)
			},
			ExpectMethod: "OPTIONS",
		},
		{
			SendReq: func(req *Request, url string) *Response {
				return req.MustPut(url)
			},
			ExpectMethod: "PUT",
		},
		{
			SendReq: func(req *Request, url string) *Response {
				return req.MustHead(url)
			},
			ExpectMethod: "HEAD",
		},
	}

	for _, tc := range testCases {
		testMethod(t, c, func(req *Request) *Response {
			return tc.SendReq(req, "/")
		}, tc.ExpectMethod, false)
	}

	// test panic
	for _, tc := range testCases {
		testMethod(t, c, func(req *Request) *Response {
			return tc.SendReq(req, "/\r\n")
		}, tc.ExpectMethod, true)
	}
}

func TestSendMethods(t *testing.T) {
	c := tc()
	testCases := []struct {
		SendReq      func(req *Request) (resp *Response, err error)
		ExpectMethod string
	}{
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Get("/")
			},
			ExpectMethod: "GET",
		},
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Post("/")
			},
			ExpectMethod: "POST",
		},
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Put("/")
			},
			ExpectMethod: "PUT",
		},
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Patch("/")
			},
			ExpectMethod: "PATCH",
		},
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Delete("/")
			},
			ExpectMethod: "DELETE",
		},
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Options("/")
			},
			ExpectMethod: "OPTIONS",
		},
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Head("/")
			},
			ExpectMethod: "HEAD",
		},
		{
			SendReq: func(req *Request) (resp *Response, err error) {
				return req.Send("GET", "/")
			},
			ExpectMethod: "GET",
		},
	}
	for _, tc := range testCases {
		testMethod(t, c, func(req *Request) *Response {
			resp, err := tc.SendReq(req)
			if err != nil {
				t.Errorf("%s %s: %s", req.method, req.RawURL, err.Error())
			}
			return resp
		}, tc.ExpectMethod, false)
	}
}

func testMethod(t *testing.T, c *Client, sendReq func(*Request) *Response, expectMethod string, expectPanic bool) {
	r := c.R()
	if expectPanic {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("Must mehod %s should panic", expectMethod)
			}
		}()
	}
	resp := sendReq(r)
	method := resp.Header.Get("Method")
	if expectMethod != method {
		t.Errorf("Expect method %s, got method %s", expectMethod, method)
	}
}

type dumpExpected struct {
	ReqHeader  bool
	ReqBody    bool
	RespHeader bool
	RespBody   bool
}

func testEnableDump(t *testing.T, fn func(r *Request) (de dumpExpected)) {
	testDump := func(c *Client) {
		r := c.R()
		de := fn(r)
		resp, err := r.SetBody(`test body`).Post("/")
		assertSuccess(t, resp, err)
		dump := resp.Dump()
		assertContains(t, dump, "user-agent", de.ReqHeader)
		assertContains(t, dump, "test body", de.ReqBody)
		assertContains(t, dump, "date", de.RespHeader)
		assertContains(t, dump, "testpost: text response", de.RespBody)
	}
	c := tc()
	testDump(c)
	testDump(c.EnableForceHTTP1())
}

func TestEnableDump(t *testing.T) {
	testCases := []func(r *Request) (d dumpExpected){
		func(r *Request) (de dumpExpected) {
			r.EnableDump()
			de.ReqHeader = true
			de.ReqBody = true
			de.RespHeader = true
			de.RespBody = true
			return
		},
		func(r *Request) (de dumpExpected) {
			r.EnableDumpWithoutHeader()
			de.ReqBody = true
			de.RespBody = true
			return
		},
		func(r *Request) (de dumpExpected) {
			r.EnableDumpWithoutBody()
			de.ReqHeader = true
			de.RespHeader = true
			return
		},
		func(r *Request) (de dumpExpected) {
			r.EnableDumpWithoutRequest()
			de.RespHeader = true
			de.RespBody = true
			return
		},
		func(r *Request) (de dumpExpected) {
			r.EnableDumpWithoutRequestBody()
			de.ReqHeader = true
			de.RespHeader = true
			de.RespBody = true
			return
		},
		func(r *Request) (de dumpExpected) {
			r.EnableDumpWithoutResponse()
			de.ReqHeader = true
			de.ReqBody = true
			return
		},
		func(r *Request) (de dumpExpected) {
			r.EnableDumpWithoutResponseBody()
			de.ReqHeader = true
			de.ReqBody = true
			de.RespHeader = true
			return
		},
		func(r *Request) (de dumpExpected) {
			r.SetDumpOptions(&DumpOptions{
				RequestHeader: true,
				RequestBody:   true,
				ResponseBody:  true,
			}).EnableDump()
			de.ReqHeader = true
			de.ReqBody = true
			de.RespBody = true
			return
		},
	}
	for _, fn := range testCases {
		testEnableDump(t, fn)
	}
}

func TestEnableDumpTo(t *testing.T) {
	buff := new(bytes.Buffer)
	resp, err := tc().R().EnableDumpTo(buff).Get("/")
	assertSuccess(t, resp, err)
	assertEqual(t, true, buff.Len() > 0)
}

func TestEnableDumpToFIle(t *testing.T) {
	tmpFile := "tmp_dumpfile_req"
	resp, err := tc().R().EnableDumpToFile(tests.GetTestFilePath(tmpFile)).Get("/")
	assertSuccess(t, resp, err)
	assertEqual(t, true, len(getTestFileContent(t, tmpFile)) > 0)
	os.Remove(tests.GetTestFilePath(tmpFile))
}

func TestBadRequest(t *testing.T) {
	resp, err := tc().R().Get("/bad-request")
	assertStatus(t, resp, err, http.StatusBadRequest, "400 Bad Request")
}

func TestSetBodyMarshal(t *testing.T) {
	username := "imroc"
	type User struct {
		Username string `json:"username" xml:"username"`
	}

	assertUsernameJson := func(body []byte) {
		var user User
		err := json.Unmarshal(body, &user)
		assertNoError(t, err)
		assertEqual(t, username, user.Username)
	}
	assertUsernameXml := func(body []byte) {
		var user User
		err := xml.Unmarshal(body, &user)
		assertNoError(t, err)
		assertEqual(t, username, user.Username)
	}

	testCases := []struct {
		Set    func(r *Request)
		Assert func(body []byte)
	}{
		{ // SetBody with map
			Set: func(r *Request) {
				m := map[string]interface{}{
					"username": username,
				}
				r.SetBody(&m)
			},
			Assert: assertUsernameJson,
		},
		{ // SetBody with struct
			Set: func(r *Request) {
				var user User
				user.Username = username
				r.SetBody(&user)
			},
			Assert: assertUsernameJson,
		},
		{ // SetBody with struct use xml
			Set: func(r *Request) {
				var user User
				user.Username = username
				r.SetBody(&user).SetContentType(xmlContentType)
			},
			Assert: assertUsernameXml,
		},
		{ // SetBodyJsonMarshal with map
			Set: func(r *Request) {
				m := map[string]interface{}{
					"username": username,
				}
				r.SetBodyJsonMarshal(&m)
			},
			Assert: assertUsernameJson,
		},
		{ // SetBodyJsonMarshal with struct
			Set: func(r *Request) {
				var user User
				user.Username = username
				r.SetBodyJsonMarshal(&user)
			},
			Assert: assertUsernameJson,
		},
		{ // SetBodyXmlMarshal with struct
			Set: func(r *Request) {
				var user User
				user.Username = username
				r.SetBodyXmlMarshal(&user)
			},
			Assert: assertUsernameXml,
		},
	}

	c := tc()
	for _, tc := range testCases {
		r := c.R()
		tc.Set(r)
		var e Echo
		resp, err := r.SetResult(&e).Post("/echo")
		assertSuccess(t, resp, err)
		tc.Assert([]byte(e.Body))
	}
}

func TestSetBody(t *testing.T) {
	body := "hello"
	fn := func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBufferString(body)), nil
	}
	c := tc()
	testCases := []struct {
		SetBody     func(r *Request)
		ContentType string
	}{
		{
			SetBody: func(r *Request) { // SetBody with `func() (io.ReadCloser, error)`
				r.SetBody(fn)
			},
		},
		{
			SetBody: func(r *Request) { //  SetBody with GetContentFunc
				r.SetBody(GetContentFunc(fn))
			},
		},
		{
			SetBody: func(r *Request) { //  SetBody with io.ReadCloser
				r.SetBody(ioutil.NopCloser(bytes.NewBufferString(body)))
			},
		},
		{
			SetBody: func(r *Request) { //  SetBody with io.Reader
				r.SetBody(bytes.NewBufferString(body))
			},
		},
		{
			SetBody: func(r *Request) { //  SetBody with string
				r.SetBody(body)
			},
			ContentType: plainTextContentType,
		},
		{
			SetBody: func(r *Request) { // SetBody with []byte
				r.SetBody([]byte(body))
			},
			ContentType: plainTextContentType,
		},
		{
			SetBody: func(r *Request) { // SetBodyString
				r.SetBodyString(body)
			},
			ContentType: plainTextContentType,
		},
		{
			SetBody: func(r *Request) { // SetBodyBytes
				r.SetBodyBytes([]byte(body))
			},
			ContentType: plainTextContentType,
		},
		{
			SetBody: func(r *Request) { // SetBodyJsonString
				r.SetBodyJsonString(body)
			},
			ContentType: jsonContentType,
		},
		{
			SetBody: func(r *Request) { // SetBodyJsonBytes
				r.SetBodyJsonBytes([]byte(body))
			},
			ContentType: jsonContentType,
		},
		{
			SetBody: func(r *Request) { // SetBodyXmlString
				r.SetBodyXmlString(body)
			},
			ContentType: xmlContentType,
		},
		{
			SetBody: func(r *Request) { // SetBodyXmlBytes
				r.SetBodyXmlBytes([]byte(body))
			},
			ContentType: xmlContentType,
		},
	}
	for _, tc := range testCases {
		r := c.R()
		tc.SetBody(r)
		var e Echo
		resp, err := r.SetResult(&e).Post("/echo")
		assertSuccess(t, resp, err)
		assertEqual(t, tc.ContentType, e.Header.Get(hdrContentTypeKey))
		assertEqual(t, body, e.Body)
	}
}

func TestCookie(t *testing.T) {
	headers := make(http.Header)
	resp, err := tc().R().SetCookies(
		&http.Cookie{
			Name:  "cookie1",
			Value: "value1",
		},
		&http.Cookie{
			Name:  "cookie2",
			Value: "value2",
		},
	).SetResult(&headers).Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "cookie1=value1; cookie2=value2", headers.Get("Cookie"))
}

func TestSetBasicAuth(t *testing.T) {
	headers := make(http.Header)
	resp, err := tc().R().
		SetBasicAuth("imroc", "123456").
		SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "Basic aW1yb2M6MTIzNDU2", headers.Get("Authorization"))
}

func TestSetBearerAuthToken(t *testing.T) {
	token := "NGU1ZWYwZDJhNmZhZmJhODhmMjQ3ZDc4"
	headers := make(http.Header)
	resp, err := tc().R().
		SetBearerAuthToken(token).
		SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "Bearer "+token, headers.Get("Authorization"))
}

func TestHeader(t *testing.T) {
	testWithAllTransport(t, testHeader)
}

func testHeader(t *testing.T, c *Client) {
	// Set User-Agent
	customUserAgent := "My Custom User Agent"
	resp, err := c.R().SetHeader(hdrUserAgentKey, customUserAgent).Get("/user-agent")
	assertSuccess(t, resp, err)
	assertEqual(t, customUserAgent, resp.String())

	// Set custom header
	headers := make(http.Header)
	resp, err = c.R().
		SetHeader("header1", "value1").
		SetHeaders(map[string]string{
			"header2": "value2",
			"header3": "value3",
		}).SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "value1", headers.Get("header1"))
	assertEqual(t, "value2", headers.Get("header2"))
	assertEqual(t, "value3", headers.Get("header3"))
}

func TestQueryParam(t *testing.T) {
	testWithAllTransport(t, testQueryParam)
}

func testQueryParam(t *testing.T, c *Client) {
	// Set query param at client level, should be overwritten at request level
	c.SetCommonQueryParam("key1", "client").
		SetCommonQueryParams(map[string]string{
			"key2": "client",
			"key3": "client",
		}).
		SetCommonQueryString("key4=client&key5=client").
		AddCommonQueryParam("key5", "extra")

	// SetQueryParam
	resp, err := c.R().
		SetQueryParam("key1", "value1").
		SetQueryParam("key2", "value2").
		SetQueryParam("key3", "value3").
		Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "key1=value1&key2=value2&key3=value3&key4=client&key5=client&key5=extra", resp.String())

	// SetQueryString
	resp, err = c.R().
		SetQueryString("key1=value1&key2=value2&key3=value3").
		Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "key1=value1&key2=value2&key3=value3&key4=client&key5=client&key5=extra", resp.String())

	// SetQueryParams
	resp, err = c.R().
		SetQueryParams(map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}).
		Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "key1=value1&key2=value2&key3=value3&key4=client&key5=client&key5=extra", resp.String())

	// SetQueryParam & SetQueryParams & SetQueryString
	resp, err = c.R().
		SetQueryParam("key1", "value1").
		SetQueryParams(map[string]string{
			"key2": "value2",
			"key3": "value3",
		}).
		SetQueryString("key4=value4&key5=value5").
		Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "key1=value1&key2=value2&key3=value3&key4=value4&key5=value5", resp.String())

	// Set same param to override
	resp, err = c.R().
		SetQueryParam("key1", "value1").
		SetQueryParams(map[string]string{
			"key2": "value2",
			"key3": "value3",
		}).
		SetQueryString("key4=value4&key5=value5").
		SetQueryParam("key1", "value11").
		SetQueryParam("key2", "value22").
		SetQueryParam("key4", "value44").
		Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "key1=value11&key2=value22&key3=value3&key4=value44&key5=value5", resp.String())

	// Add same param without override
	resp, err = c.R().
		SetQueryParam("key1", "value1").
		SetQueryParams(map[string]string{
			"key2": "value2",
			"key3": "value3",
		}).
		SetQueryString("key4=value4&key5=value5").
		AddQueryParam("key1", "value11").
		AddQueryParam("key2", "value22").
		AddQueryParam("key4", "value44").
		Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "key1=value1&key1=value11&key2=value2&key2=value22&key3=value3&key4=value4&key4=value44&key5=value5", resp.String())
}

func TestPathParam(t *testing.T) {
	testPathParam(t, tc())
	testPathParam(t, tc().EnableForceHTTP1())
}

func testPathParam(t *testing.T, c *Client) {
	username := "imroc"
	resp, err := c.R().
		SetPathParam("username", username).
		Get("/user/{username}/profile")
	assertSuccess(t, resp, err)
	assertEqual(t, fmt.Sprintf("%s's profile", username), resp.String())
}

func TestSuccess(t *testing.T) {
	testWithAllTransport(t, testSuccess)
}

func testSuccess(t *testing.T, c *Client) {
	var userInfo UserInfo
	resp, err := c.R().
		SetQueryParam("username", "imroc").
		SetResult(&userInfo).
		Get("/search")
	assertSuccess(t, resp, err)
	assertEqual(t, "roc@imroc.cc", userInfo.Email)

	userInfo = UserInfo{}
	resp, err = c.R().
		SetQueryParam("username", "imroc").
		SetQueryParam("type", "xml"). // auto unmarshal to xml
		SetResult(&userInfo).EnableDump().
		Get("/search")
	assertSuccess(t, resp, err)
	assertEqual(t, "roc@imroc.cc", userInfo.Email)
}

func TestError(t *testing.T) {
	testWithAllTransport(t, testError)
}

func testError(t *testing.T, c *Client) {
	var errMsg ErrorMessage
	resp, err := c.R().
		SetQueryParam("username", "").
		SetError(&errMsg).
		Get("/search")
	assertIsError(t, resp, err)
	assertEqual(t, 10000, errMsg.ErrorCode)

	errMsg = ErrorMessage{}
	resp, err = c.R().
		SetQueryParam("username", "test").
		SetError(&errMsg).
		Get("/search")
	assertIsError(t, resp, err)
	assertEqual(t, 10001, errMsg.ErrorCode)

	errMsg = ErrorMessage{}
	resp, err = c.R().
		SetQueryParam("username", "test").
		SetQueryParam("type", "xml"). // auto unmarshal to xml
		SetError(&errMsg).
		Get("/search")
	assertIsError(t, resp, err)
	assertEqual(t, 10001, errMsg.ErrorCode)
}

func TestForm(t *testing.T) {
	testWithAllTransport(t, testForm)
}

func testForm(t *testing.T, c *Client) {
	var userInfo UserInfo
	resp, err := c.R().
		SetFormData(map[string]string{
			"username": "imroc",
			"type":     "xml",
		}).
		SetResult(&userInfo).
		Post("/search")
	assertSuccess(t, resp, err)
	assertEqual(t, "roc@imroc.cc", userInfo.Email)

	v := make(url.Values)
	v.Add("username", "imroc")
	v.Add("type", "xml")
	resp, err = c.R().
		SetFormDataFromValues(v).
		SetResult(&userInfo).
		Post("/search")
	assertSuccess(t, resp, err)
	assertEqual(t, "roc@imroc.cc", userInfo.Email)
}

func TestHostHeaderOverride(t *testing.T) {
	testWithAllTransport(t, testHostHeaderOverride)
}

func testHostHeaderOverride(t *testing.T, c *Client) {
	resp, err := c.R().SetHeader("Host", "testhostname").Get("/host-header")
	assertSuccess(t, resp, err)
	assertEqual(t, "testhostname", resp.String())
}

func assertTraceInfo(t *testing.T, resp *Response, enable bool) {
	ti := resp.TraceInfo()
	assertEqual(t, true, resp.TotalTime() > 0)
	if !enable {
		assertEqual(t, false, ti.TotalTime > 0)
		assertIsNil(t, ti.RemoteAddr)
		assertContains(t, ti.String(), "not enabled", true)
		assertContains(t, ti.Blame(), "not enabled", true)
		return
	}

	assertContains(t, ti.String(), "not enabled", false)
	assertContains(t, ti.Blame(), "not enabled", false)
	assertEqual(t, true, ti.TotalTime > 0)
	assertEqual(t, true, ti.ConnectTime > 0)
	assertEqual(t, true, ti.FirstResponseTime > 0)
	assertEqual(t, true, ti.ResponseTime > 0)
	assertNotNil(t, ti.RemoteAddr)
	if ti.IsConnReused {
		assertEqual(t, true, ti.TCPConnectTime == 0)
		assertEqual(t, true, ti.TLSHandshakeTime == 0)
	} else {
		assertEqual(t, true, ti.TCPConnectTime > 0)
		assertEqual(t, true, ti.TLSHandshakeTime > 0)
	}
}

func assertEnableTraceInfo(t *testing.T, resp *Response) {
	assertTraceInfo(t, resp, true)
}

func assertDisableTraceInfo(t *testing.T, resp *Response) {
	assertTraceInfo(t, resp, false)
}

func TestTraceInfo(t *testing.T) {
	testWithAllTransport(t, testTraceInfo)
}

func testTraceInfo(t *testing.T, c *Client) {
	// enable trace at client level
	c.EnableTraceAll()
	resp, err := c.R().Get("/")
	assertSuccess(t, resp, err)
	assertEnableTraceInfo(t, resp)

	// disable trace at client level
	c.DisableTraceAll()
	resp, err = c.R().Get("/")
	assertSuccess(t, resp, err)
	assertDisableTraceInfo(t, resp)

	// enable trace at request level
	resp, err = c.R().EnableTrace().Get("/")
	assertSuccess(t, resp, err)
	assertEnableTraceInfo(t, resp)
}

func TestTraceOnTimeout(t *testing.T) {
	testWithAllTransport(t, func(t *testing.T, c *Client) {
		c.EnableTraceAll().SetTimeout(100 * time.Millisecond)

		resp, err := c.R().Get("http://req-nowhere.local")
		assertNotNil(t, err)
		assertNotNil(t, resp)

		ti := resp.TraceInfo()
		assertEqual(t, true, ti.DNSLookupTime >= 0)
		assertEqual(t, true, ti.ConnectTime == 0)
		assertEqual(t, true, ti.TLSHandshakeTime == 0)
		assertEqual(t, true, ti.TCPConnectTime == 0)
		assertEqual(t, true, ti.FirstResponseTime == 0)
		assertEqual(t, true, ti.ResponseTime == 0)
		assertEqual(t, true, ti.TotalTime > 0)
		assertEqual(t, true, ti.TotalTime == resp.TotalTime())
	})
}

func TestAutoDetectRequestContentType(t *testing.T) {
	c := tc()
	resp, err := c.R().SetBody(getTestFileContent(t, "sample-image.png")).Post("/content-type")
	assertSuccess(t, resp, err)
	assertEqual(t, "image/png", resp.String())

	resp, err = c.R().SetBodyJsonString(`{"msg": "test"}`).Post("/content-type")
	assertSuccess(t, resp, err)
	assertEqual(t, jsonContentType, resp.String())

	resp, err = c.R().SetContentType(xmlContentType).SetBody(`{"msg": "test"}`).Post("/content-type")
	assertSuccess(t, resp, err)
	assertEqual(t, xmlContentType, resp.String())

	resp, err = c.R().SetBody(`<html><body><h1>hello</h1></body></html>`).Post("/content-type")
	assertSuccess(t, resp, err)
	assertEqual(t, "text/html; charset=utf-8", resp.String())

	resp, err = c.R().SetBody(`hello world`).Post("/content-type")
	assertSuccess(t, resp, err)
	assertEqual(t, plainTextContentType, resp.String())
}

func TestSetFileUploadCheck(t *testing.T) {
	c := tc()
	resp, err := c.R().SetFileUpload(FileUpload{}).Post("/multipart")
	assertErrorContains(t, err, "missing param name")
	assertErrorContains(t, err, "missing filename")
	assertErrorContains(t, err, "missing file content")
	assertEqual(t, 0, len(resp.Request.uploadFiles))
}

func TestUploadMultipart(t *testing.T) {
	m := make(map[string]interface{})
	resp, err := tc().R().
		SetFile("file", tests.GetTestFilePath("sample-image.png")).
		SetFiles(map[string]string{"file": tests.GetTestFilePath("sample-file.txt")}).
		SetFormData(map[string]string{
			"param1": "value1",
			"param2": "value2",
		}).
		SetResult(&m).
		Post("/multipart")
	assertSuccess(t, resp, err)
	assertContains(t, resp.String(), "sample-image.png", true)
	assertContains(t, resp.String(), "sample-file.txt", true)
	assertContains(t, resp.String(), "value1", true)
	assertContains(t, resp.String(), "value2", true)
}

func TestFixPragmaCache(t *testing.T) {
	resp, err := tc().EnableForceHTTP1().R().Get("/pragma")
	assertSuccess(t, resp, err)
	assertEqual(t, "no-cache", resp.Header.Get("Cache-Control"))
}

func TestSetFileBytes(t *testing.T) {
	resp := uploadTextFile(t, func(r *Request) {
		r.SetFileBytes("file", "file.txt", []byte("test"))
	})
	assertEqual(t, "test", resp.String())
}

func TestSetFileReader(t *testing.T) {
	buff := bytes.NewBufferString("test")
	resp := uploadTextFile(t, func(r *Request) {
		r.SetFileReader("file", "file.txt", buff)
	})
	assertEqual(t, "test", resp.String())

	buff = bytes.NewBufferString("test")
	resp = uploadTextFile(t, func(r *Request) {
		r.SetFileReader("file", "file.txt", ioutil.NopCloser(buff))
	})
	assertEqual(t, "test", resp.String())
}

func TestSetFile(t *testing.T) {
	filename := "sample-file.txt"
	resp := uploadTextFile(t, func(r *Request) {
		r.SetFile("file", tests.GetTestFilePath(filename))
	})
	assertEqual(t, getTestFileContent(t, filename), resp.Bytes())

	resp, err := tc().R().SetFile("file", "file-not-exists.txt").Post("/file-text")
	assertErrorContains(t, err, "no such file")
}

func TestSetFiles(t *testing.T) {
	filename := "sample-file.txt"
	resp := uploadTextFile(t, func(r *Request) {
		r.SetFiles(map[string]string{
			"file": tests.GetTestFilePath(filename),
		})
	})
	assertEqual(t, getTestFileContent(t, filename), resp.Bytes())
}

func uploadTextFile(t *testing.T, setReq func(r *Request)) *Response {
	r := tc().R()
	setReq(r)
	resp, err := r.Post("/file-text")
	assertSuccess(t, resp, err)
	return resp
}
