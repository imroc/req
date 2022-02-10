package req

import (
	"bytes"
	"testing"
)

func TestClientDump(t *testing.T) {
	c := tc()
	buf := new(bytes.Buffer)
	c.EnableDumpAllTo(buf)
	resp, err := c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutHeader().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertNotContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertNotContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutBody().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertNotContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertNotContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutRequest().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertNotContains(t, buf.String(), "POST / HTTP/1.1")
	assertNotContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutRequestBody().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertNotContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutResponse().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertNotContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertNotContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	c.EnableDumpAllWithoutResponseBody().EnableDumpAllTo(buf)
	resp, err = c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertContains(t, buf.String(), "test body")
	assertContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertNotContains(t, buf.String(), "TestPost: text response")

	c = tc()
	buf = new(bytes.Buffer)
	opt := &DumpOptions{
		RequestHeader:  true,
		RequestBody:    false,
		ResponseHeader: false,
		ResponseBody:   true,
		Output:         buf,
	}
	c.SetCommonDumpOptions(opt).EnableDumpAll()
	resp, err = c.R().SetBody("test body").Post(getTestServerURL())
	assertResponse(t, resp, err)
	assertContains(t, buf.String(), "POST / HTTP/1.1")
	assertNotContains(t, buf.String(), "test body")
	assertNotContains(t, buf.String(), "HTTP/1.1 200 OK")
	assertContains(t, buf.String(), "TestPost: text response")
}
