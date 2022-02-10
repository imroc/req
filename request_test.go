package req

import (
	"net/http"
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
		assertSucess(t, resp, err)
		dump := resp.Dump()
		assertContains(t, dump, "POST / HTTP/1.1", reqHeader)
		assertContains(t, dump, "test body", reqBody)
		assertContains(t, dump, "HTTP/1.1 200 OK", respHeader)
		assertContains(t, dump, "TestPost: text response", respBody)
	}

	opt := &DumpOptions{
		RequestHeader:  true,
		RequestBody:    false,
		ResponseHeader: false,
		ResponseBody:   true,
	}
	resp, err := tr().SetDumpOptions(opt).EnableDump().SetBody("test body").Post(getTestServerURL())
	assertSucess(t, resp, err)
	dump := resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1", true)
	assertContains(t, dump, "test body", false)
	assertContains(t, dump, "HTTP/1.1 200 OK", false)
	assertContains(t, dump, "TestPost: text response", true)
}

func TestGet(t *testing.T) {
	resp, err := tr().Get("/")
	assertSucess(t, resp, err)
	assertEqual(t, "TestGet: text response", resp.String())
}

func TestBadRequest(t *testing.T) {
	resp, err := tr().Get("/bad-request")
	assertStatus(t, resp, err, http.StatusBadRequest, "400 Bad Request")
}

func TestCustomUserAgent(t *testing.T) {
	customUserAgent := "My Custom User Agent"
	resp, err := tr().SetHeader(hdrUserAgentKey, customUserAgent).Get("/user-agent")
	assertSucess(t, resp, err)
	assertEqual(t, customUserAgent, resp.String())
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
	assertSucess(t, resp, err)
	assertEqual(t, "key1=value1&key2=value2&key3=value3&key4=client&key5=client&key5=extra", resp.String())

	// SetQueryString
	resp, err = c.R().
		SetQueryString("key1=value1&key2=value2&key3=value3").
		Get("/query-parameter")
	assertSucess(t, resp, err)
	assertEqual(t, "key1=value1&key2=value2&key3=value3&key4=client&key5=client&key5=extra", resp.String())

	// SetQueryParams
	resp, err = c.R().
		SetQueryParams(map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}).
		Get("/query-parameter")
	assertSucess(t, resp, err)
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
	assertSucess(t, resp, err)
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
	assertSucess(t, resp, err)
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
	assertSucess(t, resp, err)
	assertEqual(t, "key1=value1&key1=value11&key2=value2&key2=value22&key3=value3&key4=value4&key4=value44&key5=value5", resp.String())
}
