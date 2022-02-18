package req

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/imroc/req/v3/internal/tests"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMethods(t *testing.T) {
	testMethods(t, tc())
	testMethods(t, tc().EnableForceHTTP1())
}

func testMethods(t *testing.T, c *Client) {
	resp, err := c.R().Put("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "PUT", resp.Header.Get("Method"))
	resp = c.R().MustPut("/")
	assertEqual(t, "PUT", resp.Header.Get("Method"))

	resp, err = c.R().Patch("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "PATCH", resp.Header.Get("Method"))
	resp = c.R().MustPatch("/")
	assertEqual(t, "PATCH", resp.Header.Get("Method"))

	resp, err = c.R().Delete("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "DELETE", resp.Header.Get("Method"))
	resp = c.R().MustDelete("/")
	assertEqual(t, "DELETE", resp.Header.Get("Method"))

	resp, err = c.R().Options("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "OPTIONS", resp.Header.Get("Method"))
	resp = c.R().MustOptions("/")
	assertEqual(t, "OPTIONS", resp.Header.Get("Method"))

	resp, err = c.R().Head("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "HEAD", resp.Header.Get("Method"))
	resp = c.R().MustHead("/")
	assertEqual(t, "HEAD", resp.Header.Get("Method"))

	resp, err = c.R().Get("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "GET", resp.Header.Get("Method"))
	resp = c.R().MustGet("/")
	assertEqual(t, "GET", resp.Header.Get("Method"))

	resp, err = c.R().Post("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "POST", resp.Header.Get("Method"))
	resp = c.R().MustPost("/")
	assertEqual(t, "POST", resp.Header.Get("Method"))
}

func TestEnableDump(t *testing.T) {
	testEnableDump(t, func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
		r.EnableDump()
		*reqHeader = true
		*reqBody = true
		*respHeader = true
		*respBody = true
	})
}

func TestEnableDumpWithoutRequest(t *testing.T) {
	testEnableDump(t, func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
		r.EnableDumpWithoutRequest()
		*reqHeader = false
		*reqBody = false
		*respHeader = true
		*respBody = true
	})
}

func TestEnableDumpWithoutRequestBody(t *testing.T) {
	testEnableDump(t, func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
		r.EnableDumpWithoutRequestBody()
		*reqHeader = true
		*reqBody = false
		*respHeader = true
		*respBody = true
	})
}

func TestEnableDumpWithoutResponse(t *testing.T) {
	testEnableDump(t, func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
		r.EnableDumpWithoutResponse()
		*reqHeader = true
		*reqBody = true
		*respHeader = false
		*respBody = false
	})
}

func TestEnableDumpWithoutResponseBody(t *testing.T) {
	testEnableDump(t, func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
		r.EnableDumpWithoutResponseBody()
		*reqHeader = true
		*reqBody = true
		*respHeader = true
		*respBody = false
	})
}

func TestEnableDumpWithoutHeader(t *testing.T) {
	testEnableDump(t, func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
		r.EnableDumpWithoutHeader()
		*reqHeader = false
		*reqBody = true
		*respHeader = false
		*respBody = true
	})
}

func TestEnableDumpWithoutBody(t *testing.T) {
	testEnableDump(t, func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
		r.EnableDumpWithoutBody()
		*reqHeader = true
		*reqBody = false
		*respHeader = true
		*respBody = false
	})
}

func testEnableDump(t *testing.T, fn func(r *Request, reqHeader, reqBody, respHeader, respBody *bool)) {
	testDump := func(c *Client) {
		r := c.R()
		var reqHeader, reqBody, respHeader, respBody bool
		fn(r, &reqHeader, &reqBody, &respHeader, &respBody)
		resp, err := r.SetBody(`test body`).Post("/")
		assertSuccess(t, resp, err)
		dump := resp.Dump()
		tests.AssertContains(t, dump, "user-agent", reqHeader)
		tests.AssertContains(t, dump, "test body", reqBody)
		tests.AssertContains(t, dump, "date", respHeader)
		tests.AssertContains(t, dump, "testpost: text response", respBody)
	}
	testDump(tc())
	testDump(tc().EnableForceHTTP1())
}

func TestSetDumpOptions(t *testing.T) {
	testSetDumpOptions(t, tc())
	testSetDumpOptions(t, tc().EnableForceHTTP1())
}

func testSetDumpOptions(t *testing.T, c *Client) {
	opt := &DumpOptions{
		RequestHeader:  true,
		RequestBody:    false,
		ResponseHeader: false,
		ResponseBody:   true,
	}
	resp, err := c.R().SetDumpOptions(opt).EnableDump().SetBody("test body").Post(getTestServerURL())
	assertSuccess(t, resp, err)
	dump := resp.Dump()
	tests.AssertContains(t, dump, "user-agent", true)
	tests.AssertContains(t, dump, "test body", false)
	tests.AssertContains(t, dump, "date", false)
	tests.AssertContains(t, dump, "testpost: text response", true)
}

func TestGet(t *testing.T) {
	testGet(t, tc())
	testGet(t, tc().EnableForceHTTP1())
}

func testGet(t *testing.T, c *Client) {
	resp, err := c.R().Get("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "TestGet: text response", resp.String())
}

func TestBadRequest(t *testing.T) {
	testBadRequest(t, tc())
	testBadRequest(t, tc().EnableForceHTTP1())
}

func testBadRequest(t *testing.T, c *Client) {
	resp, err := c.R().Get("/bad-request")
	assertStatus(t, resp, err, http.StatusBadRequest, "400 Bad Request")
}

func TestSetBodyMarshal(t *testing.T) {
	testSetBodyMarshal(t, tc())
	testSetBodyMarshal(t, tc().EnableForceHTTP1())
}

func testSetBodyMarshal(t *testing.T, c *Client) {
	type User struct {
		Username string `json:"username" xml:"username"`
	}

	assertUsername := func(username string) func(e *echo) {
		return func(e *echo) {
			var user User
			err := json.Unmarshal([]byte(e.Body), &user)
			assertError(t, err)
			assertEqual(t, username, user.Username)
		}
	}
	assertUsernameXml := func(username string) func(e *echo) {
		return func(e *echo) {
			var user User
			err := xml.Unmarshal([]byte(e.Body), &user)
			assertError(t, err)
			assertEqual(t, username, user.Username)
		}
	}
	testCases := []struct {
		Set    func(r *Request)
		Assert func(e *echo)
	}{
		{ // SetBody with map
			Set: func(r *Request) {
				m := map[string]interface{}{
					"username": "imroc",
				}
				r.SetBody(&m)
			},
			Assert: assertUsername("imroc"),
		},
		{ // SetBodyJsonMarshal with map
			Set: func(r *Request) {
				m := map[string]interface{}{
					"username": "imroc",
				}
				r.SetBodyJsonMarshal(&m)
			},
			Assert: assertUsername("imroc"),
		},
		{ // SetBody with struct
			Set: func(r *Request) {
				var user User
				user.Username = "imroc"
				r.SetBody(&user)
			},
			Assert: assertUsername("imroc"),
		},
		{ // SetBody with struct use xml
			Set: func(r *Request) {
				var user User
				user.Username = "imroc"
				r.SetBody(&user).SetContentType(xmlContentType)
			},
			Assert: assertUsernameXml("imroc"),
		},
		{ // SetBodyJsonMarshal with struct
			Set: func(r *Request) {
				var user User
				user.Username = "imroc"
				r.SetBodyJsonMarshal(&user)
			},
			Assert: assertUsername("imroc"),
		},
		{ // SetBodyXmlMarshal with struct
			Set: func(r *Request) {
				var user User
				user.Username = "imroc"
				r.SetBodyXmlMarshal(&user)
			},
			Assert: assertUsernameXml("imroc"),
		},
	}

	for _, cs := range testCases {
		r := c.R()
		cs.Set(r)
		var e echo
		resp, err := r.SetResult(&e).Post("/echo")
		assertSuccess(t, resp, err)
		cs.Assert(&e)
	}
}

func TestSetBodyContent(t *testing.T) {
	testSetBodyContent(t, tc())
	testSetBodyContent(t, tc().EnableForceHTTP1())
}

func testSetBodyContent(t *testing.T, c *Client) {
	var e echo
	testBody := "test body"

	testCases := []func(r *Request){
		func(r *Request) { // SetBody with string
			r.SetBody(testBody)
		},
		func(r *Request) { // SetBody with []byte
			r.SetBody([]byte(testBody))
		},
		func(r *Request) { // SetBodyString
			r.SetBodyString(testBody)
		},
		func(r *Request) { // SetBodyBytes
			r.SetBodyBytes([]byte(testBody))
		},
	}

	for _, fn := range testCases {
		r := c.R()
		fn(r)
		var e echo
		resp, err := r.SetResult(&e).Post("/echo")
		assertSuccess(t, resp, err)
		assertEqual(t, plainTextContentType, e.Header.Get(hdrContentTypeKey))
		assertEqual(t, testBody, e.Body)
	}

	// Set Reader
	testBodyReader := strings.NewReader(testBody)
	e = echo{}
	resp, err := c.R().SetBody(testBodyReader).SetResult(&e).Post("/echo")
	assertSuccess(t, resp, err)
	assertEqual(t, testBody, e.Body)
	assertEqual(t, "", e.Header.Get(hdrContentTypeKey))
}

func TestCookie(t *testing.T) {
	testCookie(t, tc())
	testCookie(t, tc().EnableForceHTTP1())
}

func testCookie(t *testing.T, c *Client) {
	headers := make(http.Header)
	resp, err := c.R().SetCookies(
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

func TestAuth(t *testing.T) {
	testAuth(t, tc())
	testAuth(t, tc().EnableForceHTTP1())
}

func testAuth(t *testing.T, c *Client) {
	headers := make(http.Header)
	resp, err := c.R().
		SetBasicAuth("imroc", "123456").
		SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "Basic aW1yb2M6MTIzNDU2", headers.Get("Authorization"))

	token := "NGU1ZWYwZDJhNmZhZmJhODhmMjQ3ZDc4"
	headers = make(http.Header)
	resp, err = c.R().
		SetBearerAuthToken(token).
		SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "Bearer "+token, headers.Get("Authorization"))
}

func TestHeader(t *testing.T) {
	testHeader(t, tc())
	testHeader(t, tc().EnableForceHTTP1())
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
	testQueryParam(t, tc())
	testQueryParam(t, tc().EnableForceHTTP1())
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
	testSuccess(t, tc())
	testSuccess(t, tc().EnableForceHTTP1())
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
	testError(t, tc())
	testError(t, tc().EnableForceHTTP1())
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
	testForm(t, tc())
	testForm(t, tc().EnableForceHTTP1())
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
	testHostHeaderOverride(t, tc())
	testHostHeaderOverride(t, tc().EnableForceHTTP1())
}

func testHostHeaderOverride(t *testing.T, c *Client) {
	resp, err := c.R().SetHeader("Host", "testhostname").Get("/host-header")
	assertSuccess(t, resp, err)
	assertEqual(t, "testhostname", resp.String())
}

func TestTraceInfo(t *testing.T) {
	testTraceInfo(t, tc())
	testTraceInfo(t, tc().EnableForceHTTP1())
	resp, err := tc().R().Get("/")
	assertSuccess(t, resp, err)
	ti := resp.TraceInfo()
	tests.AssertContains(t, ti.String(), "not enabled", true)
	tests.AssertContains(t, ti.Blame(), "not enabled", true)

	resp, err = tc().EnableTraceAll().R().Get("/")
	assertSuccess(t, resp, err)
	ti = resp.TraceInfo()
	tests.AssertContains(t, ti.String(), "not enabled", false)
	tests.AssertContains(t, ti.Blame(), "not enabled", false)
	assertEqual(t, true, resp.TotalTime() > 0)
}

func testTraceInfo(t *testing.T, c *Client) {
	// enable trace at client level
	c.EnableTraceAll()
	resp, err := c.R().Get("/")
	assertSuccess(t, resp, err)
	ti := resp.TraceInfo()
	assertEqual(t, true, ti.TotalTime > 0)
	assertEqual(t, true, ti.TCPConnectTime > 0)
	assertEqual(t, true, ti.TLSHandshakeTime > 0)
	assertEqual(t, true, ti.ConnectTime > 0)
	assertEqual(t, true, ti.FirstResponseTime > 0)
	assertEqual(t, true, ti.ResponseTime > 0)
	assertNotNil(t, ti.RemoteAddr)

	// disable trace at client level
	c.DisableTraceAll()
	resp, err = c.R().Get("/")
	assertSuccess(t, resp, err)
	ti = resp.TraceInfo()
	assertEqual(t, false, ti.TotalTime > 0)
	assertNil(t, ti.RemoteAddr)

	// enable trace at request level
	resp, err = c.R().EnableTrace().Get("/")
	assertSuccess(t, resp, err)
	ti = resp.TraceInfo()
	assertEqual(t, true, ti.TotalTime > 0)
	assertNotNil(t, ti.RemoteAddr)
}

func TestTraceOnTimeout(t *testing.T) {
	testTraceOnTimeout(t, C())
	testTraceOnTimeout(t, C().EnableForceHTTP1())
}

func testTraceOnTimeout(t *testing.T, c *Client) {
	c.EnableTraceAll().SetTimeout(100 * time.Millisecond)

	resp, err := c.R().Get("http://req-nowhere.local")
	assertNotNil(t, err)
	assertNotNil(t, resp)

	tr := resp.TraceInfo()
	assertEqual(t, true, tr.DNSLookupTime >= 0)
	assertEqual(t, true, tr.ConnectTime == 0)
	assertEqual(t, true, tr.TLSHandshakeTime == 0)
	assertEqual(t, true, tr.TCPConnectTime == 0)
	assertEqual(t, true, tr.FirstResponseTime == 0)
	assertEqual(t, true, tr.ResponseTime == 0)
	assertEqual(t, true, tr.TotalTime > 0)
	assertEqual(t, true, tr.TotalTime == resp.TotalTime())
}

func TestAutoDetectRequestContentType(t *testing.T) {
	resp, err := tc().R().SetBody(tests.GetTestFileContent(t, "sample-image.png")).Post("/raw-upload")
	assertSuccess(t, resp, err)
	assertEqual(t, "image/png", resp.Request.Headers.Get(hdrContentTypeKey))
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
	tests.AssertContains(t, resp.String(), "sample-image.png", true)
	tests.AssertContains(t, resp.String(), "sample-file.txt", true)
	tests.AssertContains(t, resp.String(), "value1", true)
	tests.AssertContains(t, resp.String(), "value2", true)
}

func TestFixPragmaCache(t *testing.T) {
	resp, err := tc().EnableForceHTTP1().R().Get("/pragma")
	assertSuccess(t, resp, err)
	assertEqual(t, "no-cache", resp.Header.Get("Cache-Control"))
}

func TestSetFileBytes(t *testing.T) {
	resp, err := tc().R().SetFileBytes("file", "file.txt", []byte("test")).Post("/file-text")
	assertSuccess(t, resp, err)
	assertEqual(t, "test", resp.String())
}

func TestSetBodyWrapper(t *testing.T) {
	b := []byte("test")
	s := string(b)
	c := tc()

	r := c.R().SetBodyXmlString(s)
	assertEqual(t, true, len(r.body) > 0)

	r = c.R().SetBodyXmlBytes(b)
	assertEqual(t, true, len(r.body) > 0)

	r = c.R().SetBodyJsonString(s)
	assertEqual(t, true, len(r.body) > 0)

	r = c.R().SetBodyJsonBytes(b)
	assertEqual(t, true, len(r.body) > 0)
}
