package req

import (
	"net/http"
	"testing"
)

func TestRequestDump(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := tc()
	resp, err := c.R().EnableDump().SetBody(`test body`).Post(ts.URL)
	assertResponse(t, resp, err)
	dump := resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1")
	assertContains(t, dump, "test body")
	assertContains(t, dump, "HTTP/1.1 200 OK")
	assertContains(t, dump, "TestPost: text response")

	resp, err = c.R().EnableDumpWithoutRequest().SetBody(`test body`).Post(ts.URL)
	assertResponse(t, resp, err)
	dump = resp.Dump()
	assertNotContains(t, dump, "POST / HTTP/1.1")
	assertNotContains(t, dump, "test body")
	assertContains(t, dump, "HTTP/1.1 200 OK")
	assertContains(t, dump, "TestPost: text response")

	resp, err = c.R().EnableDumpWithoutRequestBody().SetBody(`test body`).Post(ts.URL)
	assertResponse(t, resp, err)
	dump = resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1")
	assertNotContains(t, dump, "test body")
	assertContains(t, dump, "HTTP/1.1 200 OK")
	assertContains(t, dump, "TestPost: text response")

	resp, err = c.R().EnableDumpWithoutResponse().SetBody(`test body`).Post(ts.URL)
	assertResponse(t, resp, err)
	dump = resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1")
	assertContains(t, dump, "test body")
	assertNotContains(t, dump, "HTTP/1.1 200 OK")
	assertNotContains(t, dump, "TestPost: text response")

	resp, err = c.R().EnableDumpWithoutResponseBody().SetBody(`test body`).Post(ts.URL)
	assertResponse(t, resp, err)
	dump = resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1")
	assertContains(t, dump, "test body")
	assertContains(t, dump, "HTTP/1.1 200 OK")
	assertNotContains(t, dump, "TestPost: text response")

	resp, err = c.R().EnableDumpWithoutHeader().SetBody(`test body`).Post(ts.URL)
	assertResponse(t, resp, err)
	dump = resp.Dump()
	assertNotContains(t, dump, "POST / HTTP/1.1")
	assertContains(t, dump, "test body")
	assertNotContains(t, dump, "HTTP/1.1 200 OK")
	assertContains(t, dump, "TestPost: text response")

	resp, err = c.R().EnableDumpWithoutBody().SetBody(`test body`).Post(ts.URL)
	assertResponse(t, resp, err)
	dump = resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1")
	assertNotContains(t, dump, "test body")
	assertContains(t, dump, "HTTP/1.1 200 OK")
	assertNotContains(t, dump, "TestPost: text response")

	opt := &DumpOptions{
		RequestHeader:  true,
		RequestBody:    false,
		ResponseHeader: false,
		ResponseBody:   true,
	}
	resp, err = c.R().SetDumpOptions(opt).EnableDump().SetBody("test body").Post(ts.URL)
	assertResponse(t, resp, err)
	dump = resp.Dump()
	assertContains(t, dump, "POST / HTTP/1.1")
	assertNotContains(t, dump, "test body")
	assertNotContains(t, dump, "HTTP/1.1 200 OK")
	assertContains(t, dump, "TestPost: text response")
}

func TestGet(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := tc()
	resp, err := c.R().Get(ts.URL)
	assertResponse(t, resp, err)
	assertEqual(t, "TestGet: text response", resp.String())

	resp, err = c.R().Get(ts.URL + "/no-content")
	assertResponse(t, resp, err)
	assertEqual(t, "", resp.String())

	resp, err = c.R().Get(ts.URL + "/json")
	assertResponse(t, resp, err)
	assertEqual(t, `{"TestGet": "JSON response"}`, resp.String())
	assertEqual(t, resp.GetContentType(), "application/json")

	resp, err = c.R().Get(ts.URL + "/json-invalid")
	assertResponse(t, resp, err)
	assertEqual(t, `TestGet: Invalid JSON`, resp.String())
	assertEqual(t, resp.GetContentType(), "application/json")

	resp, err = c.R().Get(ts.URL + "/bad-request")
	assertStatus(t, resp, err, http.StatusBadRequest, "400 Bad Request")
}
