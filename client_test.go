package req

import (
	"bytes"
	"testing"
)

func TestClientDump(t *testing.T) {
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
		c := tc()
		buf := new(bytes.Buffer)
		c.EnableDumpAllTo(buf)
		var reqHeader, reqBody, respHeader, respBody bool
		fn(c, &reqHeader, &reqBody, &respHeader, &respBody)
		resp, err := c.R().SetBody(`test body`).Post("/")
		assertSuccess(t, resp, err)
		dump := buf.String()
		assertContains(t, dump, "POST / HTTP/1.1", reqHeader)
		assertContains(t, dump, "test body", reqBody)
		assertContains(t, dump, "HTTP/1.1 200 OK", respHeader)
		assertContains(t, dump, "TestPost: text response", respBody)
	}

	c := tc()
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
	assertContains(t, buf.String(), "POST / HTTP/1.1", true)
	assertContains(t, buf.String(), "test body", false)
	assertContains(t, buf.String(), "HTTP/1.1 200 OK", false)
	assertContains(t, buf.String(), "TestPost: text response", true)
}
