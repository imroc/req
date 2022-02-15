package req

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestAutoDecode(t *testing.T) {
	c := tc().DisableAutoDecode()
	resp, err := c.R().Get("/gbk")
	assertSuccess(t, resp, err)
	assertEqual(t, toGbk("我是roc"), resp.Bytes())

	resp, err = c.EnableAutoDecode().R().Get("/gbk")
	assertSuccess(t, resp, err)
	assertEqual(t, "我是roc", resp.String())
}

func TestSetTimeout(t *testing.T) {
	timeout := 100 * time.Second
	c := tc().SetTimeout(timeout)
	assertEqual(t, timeout, c.httpClient.Timeout)
}

func TestSetLogger(t *testing.T) {
	l := createDefaultLogger()
	c := tc().SetLogger(l)
	assertEqual(t, l, c.log)

	c.SetLogger(nil)
	assertEqual(t, &disableLogger{}, c.log)
}

func TestSetScheme(t *testing.T) {
	c := tc().SetScheme("https")
	assertEqual(t, "https", c.scheme)
}

func TestDebugLog(t *testing.T) {
	c := tc().EnableDebugLog()
	assertEqual(t, true, c.DebugLog)

	c.DisableDebugLog()
	assertEqual(t, false, c.DebugLog)
}

func TestSetCommonCookies(t *testing.T) {
	headers := make(http.Header)
	resp, err := tc().SetCommonCookies(&http.Cookie{
		Name:  "test",
		Value: "test",
	}).R().SetResult(&headers).Get("/header")
	assertSuccess(t, resp, err)
	assertEqual(t, "test=test", headers.Get("Cookie"))
}

func TestSetCommonQueryString(t *testing.T) {
	resp, err := tc().SetCommonQueryString("test=test").R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "test=test", resp.String())
}

func TestSetCommonPathParams(t *testing.T) {
	c := tc().SetCommonPathParams(map[string]string{"test": "test"})
	assertNotNil(t, c.PathParams)
	assertEqual(t, "test", c.PathParams["test"])
}

func TestSetCommonPathParam(t *testing.T) {
	c := tc().SetCommonPathParam("test", "test")
	assertNotNil(t, c.PathParams)
	assertEqual(t, "test", c.PathParams["test"])
}

func TestAddCommonQueryParam(t *testing.T) {
	resp, err := tc().
		AddCommonQueryParam("test", "1").
		AddCommonQueryParam("test", "2").
		R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "test=1&test=2", resp.String())
}

func TestSetCommonQueryParam(t *testing.T) {
	resp, err := tc().SetCommonQueryParam("test", "test").R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "test=test", resp.String())
}

func TestSetCommonQueryParams(t *testing.T) {
	resp, err := tc().SetCommonQueryParams(map[string]string{"test": "test"}).R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	assertEqual(t, "test=test", resp.String())
}

func TestInsecureSkipVerify(t *testing.T) {
	c := tc().EnableInsecureSkipVerify()
	assertEqual(t, true, c.t.TLSClientConfig.InsecureSkipVerify)

	c.DisableInsecureSkipVerify()
	assertEqual(t, false, c.t.TLSClientConfig.InsecureSkipVerify)
}

func TestSetTLSClientConfig(t *testing.T) {
	config := &tls.Config{InsecureSkipVerify: true}
	c := tc().SetTLSClientConfig(config)
	assertEqual(t, config, c.t.TLSClientConfig)
}

func TestCompression(t *testing.T) {
	c := tc().DisableCompression()
	assertEqual(t, true, c.t.DisableCompression)

	c.EnableCompression()
	assertEqual(t, false, c.t.DisableCompression)
}

func TestKeepAlives(t *testing.T) {
	c := tc().DisableKeepAlives()
	assertEqual(t, true, c.t.DisableKeepAlives)

	c.EnableKeepAlives()
	assertEqual(t, false, c.t.DisableKeepAlives)
}

func TestRedirect(t *testing.T) {
	_, err := tc().SetRedirectPolicy(NoRedirectPolicy()).R().Get("/unlimited-redirect")
	assertNotNil(t, err)
	assertContains(t, err.Error(), "redirect is disabled", true)

	_, err = tc().SetRedirectPolicy(MaxRedirectPolicy(3)).R().Get("/unlimited-redirect")
	assertNotNil(t, err)
	assertContains(t, err.Error(), "stopped after 3 redirects", true)

	_, err = tc().SetRedirectPolicy(SameDomainRedirectPolicy()).R().Get("/redirect-to-other")
	assertNotNil(t, err)
	assertContains(t, err.Error(), "different domain name is not allowed", true)

	_, err = tc().SetRedirectPolicy(SameHostRedirectPolicy()).R().Get("/redirect-to-other")
	assertNotNil(t, err)
	assertContains(t, err.Error(), "different host name is not allowed", true)

	_, err = tc().SetRedirectPolicy(AllowedHostRedirectPolicy("localhost", "127.0.0.1")).R().Get("/redirect-to-other")
	assertNotNil(t, err)
	assertContains(t, err.Error(), "redirect host [dummy.local] is not allowed", true)

	_, err = tc().SetRedirectPolicy(AllowedDomainRedirectPolicy("localhost", "127.0.0.1")).R().Get("/redirect-to-other")
	assertNotNil(t, err)
	assertContains(t, err.Error(), "redirect domain [dummy.local] is not allowed", true)
}

func TestGetTLSClientConfig(t *testing.T) {
	c := tc()
	config := c.GetTLSClientConfig()
	assertEqual(t, true, c.t.TLSClientConfig != nil)
	assertEqual(t, config, c.t.TLSClientConfig)
}

func TestSetRootCertFromFile(t *testing.T) {
	c := tc().SetRootCertsFromFile(getTestFilePath("sample-root.pem"))
	assertEqual(t, true, c.t.TLSClientConfig.RootCAs != nil)
}

func TestSetRootCertFromString(t *testing.T) {
	c := tc().SetRootCertFromString(string(getTestFileContent(t, "sample-root.pem")))
	assertEqual(t, true, c.t.TLSClientConfig.RootCAs != nil)
}

func TestSetCerts(t *testing.T) {
	c := tc().SetCerts(tls.Certificate{}, tls.Certificate{})
	assertEqual(t, true, len(c.t.TLSClientConfig.Certificates) == 2)
}

func TestSetCertFromFile(t *testing.T) {
	c := tc().SetCertFromFile(
		getTestFilePath("sample-client.pem"),
		getTestFilePath("sample-client-key.pem"),
	)
	assertEqual(t, true, len(c.t.TLSClientConfig.Certificates) == 1)
}

func TestSetOutputDirectory(t *testing.T) {
	outFile := "test_output_dir"
	resp, err := tc().
		SetOutputDirectory(testDataPath).
		R().SetOutputFile(outFile).
		Get("/")
	assertSuccess(t, resp, err)
	content := string(getTestFileContent(t, outFile))
	os.Remove(getTestFilePath(outFile))
	assertEqual(t, "TestGet: text response", content)
}

func TestSetBaseURL(t *testing.T) {
	baseURL := "http://dummy-req.local/test"
	resp, _ := tc().SetTimeout(time.Nanosecond).SetBaseURL(baseURL).R().Get("/req")
	assertEqual(t, baseURL+"/req", resp.Request.RawRequest.URL.String())
}

func TestSetCommonFormDataFromValues(t *testing.T) {
	expectedForm := make(url.Values)
	gotForm := make(url.Values)
	expectedForm.Set("test", "test")
	resp, err := tc().
		SetCommonFormDataFromValues(expectedForm).
		R().SetResult(&gotForm).
		Post("/form")
	assertSuccess(t, resp, err)
	assertEqual(t, "test", gotForm.Get("test"))
}

func TestSetCommonFormData(t *testing.T) {
	form := make(url.Values)
	resp, err := tc().
		SetCommonFormData(
			map[string]string{
				"test": "test",
			}).R().
		SetResult(&form).
		Post("/form")
	assertSuccess(t, resp, err)
	assertEqual(t, "test", form.Get("test"))
}

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

	dumpFile := "tmp_test_dump_file"
	c.EnableDumpAllToFile(getTestFilePath(dumpFile))
	resp, err = c.R().SetBody("test body").Post("/")
	assertSuccess(t, resp, err)
	dump := string(getTestFileContent(t, dumpFile))
	os.Remove(getTestFilePath(dumpFile))
	assertContains(t, dump, "user-agent", true)
	assertContains(t, dump, "test body", false)
	assertContains(t, dump, "date", false)
	assertContains(t, dump, "testpost: text response", true)

	buf = new(bytes.Buffer)
	c.EnableDumpAllTo(buf).EnableDumpAllAsync()
	resp, err = c.R().SetBody("test body").Post("/")
	assertSuccess(t, resp, err)
	time.Sleep(10 * time.Millisecond)
	assertContains(t, buf.String(), "user-agent", true)
	assertContains(t, buf.String(), "test body", false)
	assertContains(t, buf.String(), "date", false)
	assertContains(t, buf.String(), "testpost: text response", true)
}
