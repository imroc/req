package req

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestClientClone(t *testing.T) {
	c1 := tc().DevMode().
		SetCommonHeader("test", "test").
		SetCommonCookies(&http.Cookie{
			Name:  "test",
			Value: "test",
		}).SetCommonQueryParam("test", "test").
		SetCommonPathParam("test", "test")

	c2 := c1.Clone()
	assertEqualStruct(t, c1, c2, false, "t", "t2", "httpClient")
}

func TestDisableAutoReadResponse(t *testing.T) {
	testDisableAutoReadResponse(t, tc())
	testDisableAutoReadResponse(t, tc().EnableForceHTTP1())
}

func testDisableAutoReadResponse(t *testing.T, c *Client) {
	c.DisableAutoReadResponse()
	resp, err := c.R().Get("/")
	assertSuccess(t, resp, err)
	assertEqual(t, "", resp.String())
	result, err := resp.ToString()
	assertError(t, err)
	assertEqual(t, "TestGet: text response", result)

	resp, err = c.R().Get("/")
	assertSuccess(t, resp, err)
	_, err = ioutil.ReadAll(resp.Body)
	assertError(t, err)
}

func TestClientDump(t *testing.T) {
	testClientDump(t, func() *Client {
		return tc()
	})
	testClientDump(t, func() *Client {
		return tc().EnableForceHTTP1()
	})
}

func testClientDump(t *testing.T, newClient func() *Client) {
	testCases := []func(r *Client, reqHeader, reqBody, respHeader, respBody *bool){
		func(r *Client, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpAll()
			*reqHeader = true
			*reqBody = true
			*respHeader = true
			*respBody = true
		},
		func(r *Client, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpAllWithoutRequest()
			*reqHeader = false
			*reqBody = false
			*respHeader = true
			*respBody = true
		},
		func(r *Client, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpAllWithoutRequestBody()
			*reqHeader = true
			*reqBody = false
			*respHeader = true
			*respBody = true
		},
		func(r *Client, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpAllWithoutResponse()
			*reqHeader = true
			*reqBody = true
			*respHeader = false
			*respBody = false
		},
		func(r *Client, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpAllWithoutResponseBody()
			*reqHeader = true
			*reqBody = true
			*respHeader = true
			*respBody = false
		},
		func(r *Client, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpAllWithoutHeader()
			*reqHeader = false
			*reqBody = true
			*respHeader = false
			*respBody = true
		},
		func(r *Client, reqHeader, reqBody, respHeader, respBody *bool) {
			r.EnableDumpAllWithoutBody()
			*reqHeader = true
			*reqBody = false
			*respHeader = true
			*respBody = false
		},
	}

	for _, fn := range testCases {
		c := newClient()
		buf := new(bytes.Buffer)
		c.EnableDumpAllTo(buf)
		var reqHeader, reqBody, respHeader, respBody bool
		fn(c, &reqHeader, &reqBody, &respHeader, &respBody)
		resp, err := c.R().SetBody(`test body`).Post("/")
		assertSuccess(t, resp, err)
		dump := buf.String()
		assertContains(t, dump, "user-agent", reqHeader)
		assertContains(t, dump, "test body", reqBody)
		assertContains(t, dump, "date", respHeader)
		assertContains(t, dump, "testpost: text response", respBody)
	}

	c := newClient()
	buf := new(bytes.Buffer)
	opt := &DumpOptions{
		RequestHeader:  true,
		RequestBody:    false,
		ResponseHeader: false,
		ResponseBody:   true,
		Output:         buf,
	}
	c.SetCommonDumpOptions(opt).EnableDumpAll()
	resp, err := c.R().SetBody("test body").Post("/")
	assertSuccess(t, resp, err)
	assertContains(t, buf.String(), "user-agent", true)
	assertContains(t, buf.String(), "test body", false)
	assertContains(t, buf.String(), "date", false)
	assertContains(t, buf.String(), "testpost: text response", true)
}
