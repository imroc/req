package req

import (
	"bytes"
	"testing"
)

func TestClientDump(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := tc()
	buf := new(bytes.Buffer)
	c.EnableDumpAllTo(buf)
	resp, err := c.R().SetBody("test body").Post(ts.URL)
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutHeader().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(ts.URL)
	assertResponse(t, resp, err)
	assertNotContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertNotContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutBody().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(ts.URL)
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertNotContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertNotContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutRequestBody().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(ts.URL)
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertNotContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutResponse().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(ts.URL)
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertNotContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertNotContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutResponseBody().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(ts.URL)
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertNotContains(t, buf.String(), "TestPost: text response")
}
