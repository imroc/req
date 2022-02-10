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
		assertResponse(t, resp, err)
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
	assertResponse(t, resp, err)
	dump := resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1", true)
	assertContains(t, dump, "test body", false)
	assertContains(t, dump, "HTTP/1.1 200 OK", false)
	assertContains(t, dump, "TestPost: text response", true)
}

func TestGet(t *testing.T) {
	resp, err := tr().Get("/")
	assertResponse(t, resp, err)
	assertEqual(t, "TestGet: text response", resp.String())
}

func TestBadRequest(t *testing.T) {
	resp, err := tr().Get("/bad-request")
	assertStatus(t, resp, err, http.StatusBadRequest, "400 Bad Request")
}
