package req

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRequestDump(t *testing.T) {
	testCases := []func(r *Request, reqHeader, reqBody, respHeader, respBody *bool){
		func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDump()
			*reqHeader = true
			*reqBody = true
			*respHeader = true
			*respBody = true
		},
		func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpWithoutRequest()
			*reqHeader = false
			*reqBody = false
			*respHeader = true
			*respBody = true
		},
		func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpWithoutRequestBody()
			*reqHeader = true
			*reqBody = false
			*respHeader = true
			*respBody = true
		},
		func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpWithoutResponse()
			*reqHeader = true
			*reqBody = true
			*respHeader = false
			*respBody = false
		},
		func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpWithoutResponseBody()
			*reqHeader = true
			*reqBody = true
			*respHeader = true
			*respBody = false
		},
		func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpWithoutHeader()
			*reqHeader = false
			*reqBody = true
			*respHeader = false
			*respBody = true
		},
		func(r *Request, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpWithoutBody()
			*reqHeader = true
			*reqBody = false
			*respHeader = true
			*respBody = false
		},
	}

	for _, fn := range testCases {
		r := tr()
		var reqHeader, reqBody, respHeader, respBody bool
		fn(r, &reqHeader, &reqBody, &respHeader, &respBody)
		resp, err := r.SetBody(`test body`).Post("/")
		assertSuccess(t, resp, err)
		dump := resp.Dump()
		assertContains(t, dump, "user-agent", reqHeader)
		assertContains(t, dump, "test body", reqBody)
		assertContains(t, dump, "date", respHeader)
		assertContains(t, dump, "testpost: text response", respBody)
	}

	opt := &DumpOptions{
		RequestHeader:  true,
		RequestBody:    false,
		ResponseHeader: false,
		ResponseBody:   true,
	}
	resp, err := tr().SetDumpOptions(opt).EnableDump().SetBody("test body").Post(getTestServerURL())
	assertSuccess(t, resp, err)
	dump := resp.Dump()
	assertContains(t, dump, "user-agent", true)
	assertContains(t, dump, "test body", false)
	assertContains(t, dump, "date", false)
	assertContains(t, dump, "testpost: text response", true)
}

func TestGet(t *testing.T) {
	resp, err := tr().Get("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "TestGet: text response", resp.String())
}

func TestBadRequest(t *testing.T) {
	resp, err := tr().Get("/bad-request")
	assertStatus(t, resp, err, http.StatusBadRequest, "400 Bad Request")
}

func TestSetBodyMarshal(t *testing.T) {
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

	for _, c := range testCases {
		r := tr()
		c.Set(r)
		var e echo
		resp, err := r.SetResult(&e).Post("/echo")
		assertSuccess(t, resp, err)
		c.Assert(&e)
	}
}

func TestSetBodyContent(t *testing.T) {
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
		r := tr()
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
	resp, err := tr().SetBody(testBodyReader).SetResult(&e).Post("/echo")
	assertSuccess(t, resp, err)
	assertEqual(t, testBody, e.Body)
	assertEqual(t, "", e.Header.Get(hdrContentTypeKey))
}

func TestCookie(t *testing.T) {
	headers := make(http.Header)
	resp, err := tr().SetCookies(
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
	headers := make(http.Header)
	resp, err := tr().
		SetBasicAuth("imroc", "123456").
		SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "Basic aW1yb2M6MTIzNDU2", headers.Get("Authorization"))

	token := "NGU1ZWYwZDJhNmZhZmJhODhmMjQ3ZDc4"
	headers = make(http.Header)
	resp, err = tr().
		SetBearerAuthToken(token).
		SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "Bearer "+token, headers.Get("Authorization"))
}

func TestHeader(t *testing.T) {
	// Set User-Agent
	customUserAgent := "My Custom User Agent"
	resp, err := tr().SetHeader(hdrUserAgentKey, customUserAgent).Get("/user-agent")
	assertSuccess(t, resp, err)
	assertEqual(t, customUserAgent, resp.String())

	// Set custom header
	headers := make(http.Header)
	resp, err = tr().
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
	c := tc()

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
	username := "imroc"
	resp, err := tr().
		SetPathParam("username", username).
		Get("/user/{username}/profile")
	assertSuccess(t, resp, err)
	assertEqual(t, fmt.Sprintf("%s's profile", username), resp.String())
}
