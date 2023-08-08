package req

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/tests"
	"golang.org/x/net/publicsuffix"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestWrapRoundTrip(t *testing.T) {
	i, j, a, b := 0, 0, 0, 0
	c := tc().WrapRoundTripFunc(func(rt RoundTripper) RoundTripFunc {
		return func(req *Request) (resp *Response, err error) {
			a = 1
			resp, err = rt.RoundTrip(req)
			b = 1
			return
		}
	})
	c.GetTransport().WrapRoundTripFunc(func(rt http.RoundTripper) HttpRoundTripFunc {
		return func(req *http.Request) (resp *http.Response, err error) {
			i = 1
			resp, err = rt.RoundTrip(req)
			j = 1
			return
		}
	})
	resp, err := c.R().Get("/")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, 1, i)
	tests.AssertEqual(t, 1, j)
	tests.AssertEqual(t, 1, a)
	tests.AssertEqual(t, 1, b)
}

func TestAllowGetMethodPayload(t *testing.T) {
	c := tc()
	resp, err := c.R().SetBody("test").Get("/payload")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test", resp.String())

	c.DisableAllowGetMethodPayload()
	resp, err = c.R().SetBody("test").Get("/payload")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "", resp.String())

	c.EnableAllowGetMethodPayload()
	resp, err = c.R().SetBody("test").Get("/payload")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test", resp.String())
}

func TestSetTLSHandshakeTimeout(t *testing.T) {
	timeout := 2 * time.Second
	c := tc().SetTLSHandshakeTimeout(timeout)
	tests.AssertEqual(t, timeout, c.TLSHandshakeTimeout)
}

func TestSetDial(t *testing.T) {
	testErr := errors.New("test")
	testDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return nil, testErr
	}
	c := tc().SetDial(testDial)
	_, err := c.DialContext(nil, "", "")
	tests.AssertEqual(t, testErr, err)
}

func TestSetDialTLS(t *testing.T) {
	testErr := errors.New("test")
	testDialTLS := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return nil, testErr
	}
	c := tc().SetDialTLS(testDialTLS)
	_, err := c.DialTLSContext(nil, "", "")
	tests.AssertEqual(t, testErr, err)
}

func TestSetFuncs(t *testing.T) {
	testErr := errors.New("test")
	marshalFunc := func(v interface{}) ([]byte, error) {
		return nil, testErr
	}
	unmarshalFunc := func(data []byte, v interface{}) error {
		return testErr
	}
	c := tc().
		SetJsonMarshal(marshalFunc).
		SetJsonUnmarshal(unmarshalFunc).
		SetXmlMarshal(marshalFunc).
		SetXmlUnmarshal(unmarshalFunc)

	_, err := c.jsonMarshal(nil)
	tests.AssertEqual(t, testErr, err)
	err = c.jsonUnmarshal(nil, nil)
	tests.AssertEqual(t, testErr, err)

	_, err = c.xmlMarshal(nil)
	tests.AssertEqual(t, testErr, err)
	err = c.xmlUnmarshal(nil, nil)
	tests.AssertEqual(t, testErr, err)
}

func TestSetCookieJar(t *testing.T) {
	c := tc().SetCookieJar(nil)
	tests.AssertEqual(t, nil, c.httpClient.Jar)
}

func TestTraceAll(t *testing.T) {
	c := tc().EnableTraceAll()
	resp, err := c.R().Get("/")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, true, resp.TraceInfo().TotalTime > 0)

	c.DisableTraceAll()
	resp, err = c.R().Get("/")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, true, resp.TraceInfo().TotalTime == 0)
}

func TestOnAfterResponse(t *testing.T) {
	c := tc()
	len1 := len(c.afterResponse)
	c.OnAfterResponse(func(client *Client, response *Response) error {
		return nil
	})
	len2 := len(c.afterResponse)
	tests.AssertEqual(t, true, len1+1 == len2)
}

func TestOnBeforeRequest(t *testing.T) {
	c := tc().OnBeforeRequest(func(client *Client, request *Request) error {
		return nil
	})
	tests.AssertEqual(t, true, len(c.udBeforeRequest) == 1)
}

func TestSetProxyURL(t *testing.T) {
	c := tc().SetProxyURL("http://dummy.proxy.local")
	u, err := c.Proxy(nil)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, "http://dummy.proxy.local", u.String())
}

func TestSetProxy(t *testing.T) {
	u, _ := url.Parse("http://dummy.proxy.local")
	proxy := http.ProxyURL(u)
	c := tc().SetProxy(proxy)
	uu, err := c.Proxy(nil)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, u.String(), uu.String())
}

func TestSetCommonContentType(t *testing.T) {
	c := tc().SetCommonContentType(header.JsonContentType)
	tests.AssertEqual(t, header.JsonContentType, c.Headers.Get(header.ContentType))
}

func TestSetCommonHeader(t *testing.T) {
	c := tc().SetCommonHeader("my-header", "my-value")
	tests.AssertEqual(t, "my-value", c.Headers.Get("my-header"))
}

func TestSetCommonHeaderNonCanonical(t *testing.T) {
	c := tc().SetCommonHeaderNonCanonical("my-Header", "my-value")
	tests.AssertEqual(t, "my-value", c.Headers["my-Header"][0])
}

func TestSetCommonHeaders(t *testing.T) {
	c := tc().SetCommonHeaders(map[string]string{
		"header1": "value1",
		"header2": "value2",
	})
	tests.AssertEqual(t, "value1", c.Headers.Get("header1"))
	tests.AssertEqual(t, "value2", c.Headers.Get("header2"))
}

func TestSetCommonHeadersNonCanonical(t *testing.T) {
	c := tc().SetCommonHeadersNonCanonical(map[string]string{
		"my-Header": "my-value",
	})
	tests.AssertEqual(t, "my-value", c.Headers["my-Header"][0])
}

func TestSetCommonBasicAuth(t *testing.T) {
	c := tc().SetCommonBasicAuth("imroc", "123456")
	tests.AssertEqual(t, "Basic aW1yb2M6MTIzNDU2", c.Headers.Get("Authorization"))
}

func TestSetCommonBearerAuthToken(t *testing.T) {
	c := tc().SetCommonBearerAuthToken("123456")
	tests.AssertEqual(t, "Bearer 123456", c.Headers.Get("Authorization"))
}

func TestSetUserAgent(t *testing.T) {
	c := tc().SetUserAgent("test")
	tests.AssertEqual(t, "test", c.Headers.Get(header.UserAgent))
}

func TestAutoDecode(t *testing.T) {
	c := tc().DisableAutoDecode()
	resp, err := c.R().Get("/gbk")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, toGbk("我是roc"), resp.Bytes())

	resp, err = c.EnableAutoDecode().R().Get("/gbk")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "我是roc", resp.String())

	resp, err = c.SetAutoDecodeContentType("html").R().Get("/gbk")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, toGbk("我是roc"), resp.Bytes())
	resp, err = c.SetAutoDecodeContentType("text").R().Get("/gbk")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "我是roc", resp.String())
	resp, err = c.SetAutoDecodeContentTypeFunc(func(contentType string) bool {
		if strings.Contains(contentType, "text") {
			return true
		}
		return false
	}).R().Get("/gbk")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "我是roc", resp.String())

	resp, err = c.SetAutoDecodeAllContentType().R().Get("/gbk-no-charset")
	assertSuccess(t, resp, err)
	tests.AssertContains(t, resp.String(), "我是roc", true)
}

func TestSetTimeout(t *testing.T) {
	timeout := 100 * time.Second
	c := tc().SetTimeout(timeout)
	tests.AssertEqual(t, timeout, c.httpClient.Timeout)
}

func TestSetLogger(t *testing.T) {
	l := createDefaultLogger()
	c := tc().SetLogger(l)
	tests.AssertEqual(t, l, c.log)

	c.SetLogger(nil)
	tests.AssertEqual(t, &disableLogger{}, c.log)
}

func TestSetScheme(t *testing.T) {
	c := tc().SetScheme("https")
	tests.AssertEqual(t, "https", c.scheme)
}

func TestDebugLog(t *testing.T) {
	c := tc().EnableDebugLog()
	tests.AssertEqual(t, true, c.DebugLog)

	c.DisableDebugLog()
	tests.AssertEqual(t, false, c.DebugLog)
}

func TestSetCommonCookies(t *testing.T) {
	headers := make(http.Header)
	resp, err := tc().SetCommonCookies(&http.Cookie{
		Name:  "test",
		Value: "test",
	}).R().SetSuccessResult(&headers).Get("/header")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test=test", headers.Get("Cookie"))
}

func TestSetCommonQueryString(t *testing.T) {
	resp, err := tc().SetCommonQueryString("test=test").R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test=test", resp.String())
}

func TestSetCommonPathParams(t *testing.T) {
	c := tc().SetCommonPathParams(map[string]string{"test": "test"})
	tests.AssertNotNil(t, c.PathParams)
	tests.AssertEqual(t, "test", c.PathParams["test"])
}

func TestSetCommonPathParam(t *testing.T) {
	c := tc().SetCommonPathParam("test", "test")
	tests.AssertNotNil(t, c.PathParams)
	tests.AssertEqual(t, "test", c.PathParams["test"])
}

func TestAddCommonQueryParam(t *testing.T) {
	resp, err := tc().
		AddCommonQueryParam("test", "1").
		AddCommonQueryParam("test", "2").
		R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test=1&test=2", resp.String())
}

func TestSetCommonQueryParam(t *testing.T) {
	resp, err := tc().SetCommonQueryParam("test", "test").R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test=test", resp.String())
}

func TestSetCommonQueryParams(t *testing.T) {
	resp, err := tc().SetCommonQueryParams(map[string]string{"test": "test"}).R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test=test", resp.String())
}

func TestInsecureSkipVerify(t *testing.T) {
	c := tc().EnableInsecureSkipVerify()
	tests.AssertEqual(t, true, c.TLSClientConfig.InsecureSkipVerify)

	c.DisableInsecureSkipVerify()
	tests.AssertEqual(t, false, c.TLSClientConfig.InsecureSkipVerify)
}

func TestSetTLSClientConfig(t *testing.T) {
	config := &tls.Config{InsecureSkipVerify: true}
	c := tc().SetTLSClientConfig(config)
	tests.AssertEqual(t, config, c.TLSClientConfig)
}

func TestCompression(t *testing.T) {
	c := tc().DisableCompression()
	tests.AssertEqual(t, true, c.Transport.DisableCompression)

	c.EnableCompression()
	tests.AssertEqual(t, false, c.Transport.DisableCompression)
}

func TestKeepAlives(t *testing.T) {
	c := tc().DisableKeepAlives()
	tests.AssertEqual(t, true, c.Transport.DisableKeepAlives)

	c.EnableKeepAlives()
	tests.AssertEqual(t, false, c.Transport.DisableKeepAlives)
}

func TestRedirect(t *testing.T) {
	_, err := tc().SetRedirectPolicy(NoRedirectPolicy()).R().Get("/unlimited-redirect")
	tests.AssertIsNil(t, err)

	_, err = tc().SetRedirectPolicy(MaxRedirectPolicy(3)).R().Get("/unlimited-redirect")
	tests.AssertNotNil(t, err)
	tests.AssertContains(t, err.Error(), "stopped after 3 redirects", true)

	_, err = tc().SetRedirectPolicy(SameDomainRedirectPolicy()).R().Get("/redirect-to-other")
	tests.AssertNotNil(t, err)
	tests.AssertContains(t, err.Error(), "different domain name is not allowed", true)

	_, err = tc().SetRedirectPolicy(SameHostRedirectPolicy()).R().Get("/redirect-to-other")
	tests.AssertNotNil(t, err)
	tests.AssertContains(t, err.Error(), "different host name is not allowed", true)

	_, err = tc().SetRedirectPolicy(AllowedHostRedirectPolicy("localhost", "127.0.0.1")).R().Get("/redirect-to-other")
	tests.AssertNotNil(t, err)
	tests.AssertContains(t, err.Error(), "redirect host [dummy.local] is not allowed", true)

	_, err = tc().SetRedirectPolicy(AllowedDomainRedirectPolicy("localhost", "127.0.0.1")).R().Get("/redirect-to-other")
	tests.AssertNotNil(t, err)
	tests.AssertContains(t, err.Error(), "redirect domain [dummy.local] is not allowed", true)

	c := tc().SetRedirectPolicy(AlwaysCopyHeaderRedirectPolicy("Authorization"))
	newHeader := make(http.Header)
	oldHeader := make(http.Header)
	oldHeader.Set("Authorization", "test")
	c.GetClient().CheckRedirect(&http.Request{
		Header: newHeader,
	}, []*http.Request{&http.Request{
		Header: oldHeader,
	}})
	tests.AssertEqual(t, "test", newHeader.Get("Authorization"))
}

func TestGetTLSClientConfig(t *testing.T) {
	c := tc()
	config := c.GetTLSClientConfig()
	tests.AssertEqual(t, true, c.TLSClientConfig != nil)
	tests.AssertEqual(t, config, c.TLSClientConfig)
}

func TestSetRootCertFromFile(t *testing.T) {
	c := tc().SetRootCertsFromFile(tests.GetTestFilePath("sample-root.pem"))
	tests.AssertEqual(t, true, c.TLSClientConfig.RootCAs != nil)
}

func TestSetRootCertFromString(t *testing.T) {
	c := tc().SetRootCertFromString(string(getTestFileContent(t, "sample-root.pem")))
	tests.AssertEqual(t, true, c.TLSClientConfig.RootCAs != nil)
}

func TestSetCerts(t *testing.T) {
	c := tc().SetCerts(tls.Certificate{}, tls.Certificate{})
	tests.AssertEqual(t, true, len(c.TLSClientConfig.Certificates) == 2)
}

func TestSetCertFromFile(t *testing.T) {
	c := tc().SetCertFromFile(
		tests.GetTestFilePath("sample-client.pem"),
		tests.GetTestFilePath("sample-client-key.pem"),
	)
	tests.AssertEqual(t, true, len(c.TLSClientConfig.Certificates) == 1)
}

func TestSetOutputDirectory(t *testing.T) {
	outFile := "test_output_dir"
	resp, err := tc().
		SetOutputDirectory(testDataPath).
		R().SetOutputFile(outFile).
		Get("/")
	assertSuccess(t, resp, err)
	content := string(getTestFileContent(t, outFile))
	os.Remove(tests.GetTestFilePath(outFile))
	tests.AssertEqual(t, "TestGet: text response", content)
}

func TestSetBaseURL(t *testing.T) {
	baseURL := "http://dummy-req.local/test"
	resp, _ := tc().SetTimeout(time.Nanosecond).SetBaseURL(baseURL).R().Get("/req")
	tests.AssertEqual(t, baseURL+"/req", resp.Request.RawRequest.URL.String())
}

func TestSetCommonFormDataFromValues(t *testing.T) {
	expectedForm := make(url.Values)
	gotForm := make(url.Values)
	expectedForm.Set("test", "test")
	resp, err := tc().
		SetCommonFormDataFromValues(expectedForm).
		R().SetSuccessResult(&gotForm).
		Post("/form")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test", gotForm.Get("test"))
}

func TestSetCommonFormData(t *testing.T) {
	form := make(url.Values)
	resp, err := tc().
		SetCommonFormData(
			map[string]string{
				"test": "test",
			}).R().
		SetSuccessResult(&form).
		Post("/form")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test", form.Get("test"))
}

func TestClientClone(t *testing.T) {
	c1 := tc().DevMode().
		SetCommonHeader("test", "test").
		SetCommonCookies(&http.Cookie{
			Name:  "test",
			Value: "test",
		}).SetCommonQueryParam("test", "test").
		SetCommonPathParam("test", "test").
		SetCommonRetryCount(2).
		SetCommonFormData(map[string]string{"test": "test"}).
		OnBeforeRequest(func(c *Client, r *Request) error { return nil })

	c2 := c1.Clone()
	assertClone(t, c1, c2)
}

func TestDisableAutoReadResponse(t *testing.T) {
	testWithAllTransport(t, testDisableAutoReadResponse)
}

func testDisableAutoReadResponse(t *testing.T, c *Client) {
	c.DisableAutoReadResponse()
	resp, err := c.R().Get("/")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "", resp.String())
	result, err := resp.ToString()
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, "TestGet: text response", result)

	resp, err = c.R().Get("/")
	assertSuccess(t, resp, err)
	_, err = io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)
}

func testEnableDumpAll(t *testing.T, fn func(c *Client) (de dumpExpected)) {
	testDump := func(c *Client) {
		buff := new(bytes.Buffer)
		c.EnableDumpAllTo(buff)
		r := c.R()
		de := fn(c)
		resp, err := r.SetBody(`test body`).Post("/")
		assertSuccess(t, resp, err)
		dump := buff.String()
		tests.AssertContains(t, dump, "user-agent", de.ReqHeader)
		tests.AssertContains(t, dump, "test body", de.ReqBody)
		tests.AssertContains(t, dump, "date", de.RespHeader)
		tests.AssertContains(t, dump, "testpost: text response", de.RespBody)
	}
	c := tc()
	testDump(c)
	testDump(c.EnableForceHTTP1())
}

func TestEnableDumpAll(t *testing.T) {
	testCases := []func(c *Client) (d dumpExpected){
		func(c *Client) (de dumpExpected) {
			c.EnableDumpAll()
			de.ReqHeader = true
			de.ReqBody = true
			de.RespHeader = true
			de.RespBody = true
			return
		},
		func(c *Client) (de dumpExpected) {
			c.EnableDumpAllWithoutHeader()
			de.ReqBody = true
			de.RespBody = true
			return
		},
		func(c *Client) (de dumpExpected) {
			c.EnableDumpAllWithoutBody()
			de.ReqHeader = true
			de.RespHeader = true
			return
		},
		func(c *Client) (de dumpExpected) {
			c.EnableDumpAllWithoutRequest()
			de.RespHeader = true
			de.RespBody = true
			return
		},
		func(c *Client) (de dumpExpected) {
			c.EnableDumpAllWithoutRequestBody()
			de.ReqHeader = true
			de.RespHeader = true
			de.RespBody = true
			return
		},
		func(c *Client) (de dumpExpected) {
			c.EnableDumpAllWithoutResponse()
			de.ReqHeader = true
			de.ReqBody = true
			return
		},
		func(c *Client) (de dumpExpected) {
			c.EnableDumpAllWithoutResponseBody()
			de.ReqHeader = true
			de.ReqBody = true
			de.RespHeader = true
			return
		},
		func(c *Client) (de dumpExpected) {
			c.SetCommonDumpOptions(&DumpOptions{
				RequestHeader: true,
				RequestBody:   true,
				ResponseBody:  true,
			}).EnableDumpAll()
			de.ReqHeader = true
			de.ReqBody = true
			de.RespBody = true
			return
		},
	}
	for _, fn := range testCases {
		testEnableDumpAll(t, fn)
	}
}

func TestEnableDumpAllToFile(t *testing.T) {
	c := tc()
	dumpFile := "tmp_test_dump_file"
	c.EnableDumpAllToFile(tests.GetTestFilePath(dumpFile))
	resp, err := c.R().SetBody("test body").Post("/")
	assertSuccess(t, resp, err)
	dump := string(getTestFileContent(t, dumpFile))
	os.Remove(tests.GetTestFilePath(dumpFile))
	tests.AssertContains(t, dump, "user-agent", true)
	tests.AssertContains(t, dump, "test body", true)
	tests.AssertContains(t, dump, "date", true)
	tests.AssertContains(t, dump, "testpost: text response", true)
}

func TestEnableDumpAllAsync(t *testing.T) {
	c := tc()
	buf := new(bytes.Buffer)
	c.EnableDumpAllTo(buf).EnableDumpAllAsync()
	tests.AssertEqual(t, true, c.getDumpOptions().Async)
}

func TestSetResponseBodyTransformer(t *testing.T) {
	c := tc().SetResponseBodyTransformer(func(rawBody []byte, req *Request, resp *Response) (transformedBody []byte, err error) {
		if resp.IsSuccessState() {
			result, err := url.QueryUnescape(string(rawBody))
			return []byte(result), err
		}
		return rawBody, nil
	})
	user := &UserInfo{}
	resp, err := c.R().SetSuccessResult(user).Get("/urlencode")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, user.Username, "我是roc")
	tests.AssertEqual(t, user.Email, "roc@imroc.cc")
}

func TestSetResultStateCheckFunc(t *testing.T) {
	c := tc().SetResultStateCheckFunc(func(resp *Response) ResultState {
		if resp.StatusCode == http.StatusOK {
			return SuccessState
		} else {
			return ErrorState
		}
	})
	resp, err := c.R().Get("/status?code=200")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, SuccessState, resp.ResultState())

	resp, err = c.R().Get("/status?code=201")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, ErrorState, resp.ResultState())

	resp, err = c.R().Get("/status?code=399")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, ErrorState, resp.ResultState())

	resp, err = c.R().Get("/status?code=404")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, ErrorState, resp.ResultState())
}
func TestCloneCookieJar(t *testing.T) {
	c1 := C()
	c2 := c1.Clone()
	tests.AssertEqual(t, true, c1.httpClient.Jar != c2.httpClient.Jar)

	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	c1.SetCookieJar(jar)
	c2 = c1.Clone()
	tests.AssertEqual(t, true, c1.httpClient.Jar == c2.httpClient.Jar)

	c2.SetCookieJar(nil)
	tests.AssertEqual(t, true, c2.cookiejarFactory == nil)
	tests.AssertEqual(t, true, c2.httpClient.Jar == nil)
}
