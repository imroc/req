package req

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/tests"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/publicsuffix"
)

func TestRetryCancelledContext(t *testing.T) {
	cancelledCtx, done := context.WithCancel(context.Background())
	done()

	client := tc().
		SetCommonRetryCount(2).
		SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second)

	res, err := client.R().SetContext(cancelledCtx).Get("/")

	tests.AssertEqual(t, 0, res.Request.RetryAttempt)
	tests.AssertNotNil(t, err)
	tests.AssertErrorContains(t, err, "context canceled")
}

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

func TestSetResolver(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	// Broken custom resolver: hostname lookups fail, but IP literals still work.
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, errors.New("custom resolver down")
		},
	}
	c := C().SetResolver(r).EnableInsecureSkipVerify()
	tests.AssertNotNil(t, c.DialContext)

	resp, err := c.R().Get(ts.URL)
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "ok", resp.String())

	// Hostname must go through the custom resolver and fail.
	_, err = c.SetTimeout(3 * time.Second).R().Get("https://example.invalid/")
	tests.AssertNotNil(t, err)
	tests.AssertErrorContains(t, err, "custom resolver down")

	// nil resolver uses the default resolver (still installs a dialer).
	c2 := C().SetResolver(nil)
	tests.AssertNotNil(t, c2.DialContext)
}

func TestSetHosts(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hosts-ok"))
	}))
	defer ts.Close()

	ip, port, err := net.SplitHostPort(ts.Listener.Addr().String())
	tests.AssertNoError(t, err)

	c := C().
		EnableInsecureSkipVerify().
		EnableForceHTTP1().
		SetHosts(map[string]string{
			"Custom.Example": ip, // case-insensitive match
		})

	resp, err := c.R().Get(fmt.Sprintf("https://custom.example:%s/", port))
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "hosts-ok", resp.String())
}

func TestSetHostsHTTP2(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("h2-ok"))
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	ip, port, err := net.SplitHostPort(ts.Listener.Addr().String())
	tests.AssertNoError(t, err)

	c := C().
		EnableInsecureSkipVerify().
		EnableForceHTTP2().
		SetHosts(map[string]string{
			"h2.example": ip,
		})

	resp, err := c.R().Get(fmt.Sprintf("https://h2.example:%s/", port))
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "h2-ok", resp.String())
	tests.AssertEqual(t, "HTTP/2.0", resp.Proto)
}

func TestSetHostsUnknownHostFailsFast(t *testing.T) {
	c := C().
		EnableInsecureSkipVerify().
		SetTimeout(5 * time.Second).
		SetHosts(map[string]string{
			"known.example": "127.0.0.1",
		})

	start := time.Now()
	_, err := c.R().Get("https://unknown.example/")
	elapsed := time.Since(start)

	tests.AssertNotNil(t, err)
	tests.AssertErrorContains(t, err, "no such host")
	// Must not wait on system DNS timeouts.
	if elapsed > 2*time.Second {
		t.Fatalf("unknown host took too long: %v", elapsed)
	}
}

func TestSetHostsEmptyMapFailsFast(t *testing.T) {
	for _, hosts := range []map[string]string{nil, {}} {
		c := C().SetHosts(hosts).SetTimeout(2 * time.Second)
		start := time.Now()
		_, err := c.DialContext(context.Background(), "tcp", "any.example:80")
		elapsed := time.Since(start)
		tests.AssertNotNil(t, err)
		tests.AssertErrorContains(t, err, "no such host")
		if elapsed > time.Second {
			t.Fatalf("empty hosts map dial took too long: %v (hosts=%v)", elapsed, hosts)
		}
	}
}

func TestSetHostsIgnoresCallerMapMutation(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	ip, port, err := net.SplitHostPort(ts.Listener.Addr().String())
	tests.AssertNoError(t, err)

	hosts := map[string]string{"static.example": ip}
	c := C().EnableInsecureSkipVerify().EnableForceHTTP1().SetHosts(hosts)
	// Mutating the original map after SetHosts must not affect the client.
	delete(hosts, "static.example")
	hosts["static.example"] = "0.0.0.0"

	resp, err := c.R().Get(fmt.Sprintf("https://static.example:%s/", port))
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "ok", resp.String())
}

func TestSetHostsReplacedBySetDial(t *testing.T) {
	called := false
	c := C().
		SetTimeout(2 * time.Second).
		SetHosts(map[string]string{"x.example": "127.0.0.1"}).
		SetDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
			called = true
			return nil, errors.New("custom dial")
		})

	_, err := c.R().Get("https://x.example/")
	tests.AssertNotNil(t, err)
	tests.AssertEqual(t, true, called)
	tests.AssertErrorContains(t, err, "custom dial")
}

func TestSetHostsIPLiteralPassthrough(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("literal-ok"))
	}))
	defer ts.Close()

	// Map does not include the server IP; IP-literal URLs must still work.
	c := C().
		EnableInsecureSkipVerify().
		SetHosts(map[string]string{
			"only.example": "10.0.0.1",
		})

	resp, err := c.R().Get(ts.URL)
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "literal-ok", resp.String())
}

func TestSetHostsScopedIPv6LiteralPassthrough(t *testing.T) {
	c := C().SetHosts(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.DialContext(ctx, "tcp", "[fe80::1%eth0]:443")
	tests.AssertNotNil(t, err)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("scoped IPv6 literal was not dialed directly: %v", err)
	}
}

func TestSetHostsRejectsProxy(t *testing.T) {
	proxyURL := "http://127.0.0.1:1"
	cases := []struct {
		name   string
		client *Client
	}{
		{
			name: "proxy configured before SetHosts",
			client: C().
				SetProxyURL(proxyURL).
				SetHosts(map[string]string{"allowed.example": "127.0.0.1"}),
		},
		{
			name: "proxy configured after SetHosts",
			client: C().
				SetHosts(map[string]string{"allowed.example": "127.0.0.1"}).
				SetProxyURL(proxyURL),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			for _, target := range []string{"http://allowed.example/", "http://unknown.example/"} {
				_, err := tt.client.R().Get(target)
				tests.AssertNotNil(t, err)
				tests.AssertErrorContains(t, err, "SetHosts cannot be used with a proxy")
			}
		})
	}
}

func TestSetHostsInvalidIPNoDNS(t *testing.T) {
	c := C().SetHosts(map[string]string{
		"bad.example": "not-an-ip",
	})

	start := time.Now()
	_, err := c.DialContext(context.Background(), "tcp", "bad.example:80")
	elapsed := time.Since(start)

	tests.AssertNotNil(t, err)
	tests.AssertErrorContains(t, err, "invalid IP address")
	tests.AssertErrorContains(t, err, "not-an-ip")
	if elapsed > time.Second {
		t.Fatalf("invalid IP dial took too long (possible DNS fallback): %v", elapsed)
	}
}

func TestSetHostsIPv6BracketedValue(t *testing.T) {
	// Bracketed values must normalize to a single JoinHostPort form, not [[::1]]:port.
	ip := net.ParseIP(strings.Trim("[::1]", "[]"))
	tests.AssertNotNil(t, ip)
	tests.AssertEqual(t, "[::1]:443", net.JoinHostPort(ip.String(), "443"))

	// Dial with bracketed and plain map values must not produce address-parse errors.
	// Connection may fail (no listener / no IPv6), but the address form must be valid.
	for _, value := range []string{"[::1]", "::1"} {
		c := C().SetHosts(map[string]string{"v6.example": value})
		_, err := c.DialContext(context.Background(), "tcp", "v6.example:1")
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "too many colons") ||
				strings.Contains(msg, "invalid IP") ||
				strings.Contains(msg, "[[") {
				t.Fatalf("value %q produced bad address form: %v", value, err)
			}
		}
	}

	// If IPv6 loopback is available, do a full connect as well.
	ln, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Logf("skipping live IPv6 connect: %v", err)
		return
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	tests.AssertNoError(t, err)
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()
	c := C().SetHosts(map[string]string{"v6.example": "[::1]"})
	conn, err := c.DialContext(context.Background(), "tcp", net.JoinHostPort("v6.example", port))
	tests.AssertNoError(t, err)
	_ = conn.Close()
}

func TestHostsMapKeyIDNA(t *testing.T) {
	// Non-ASCII host should normalize to the same key as its punycode form.
	unicodeHost := "bücher.example"
	asciiHost, err := idnaASCII(unicodeHost)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, hostsMapKey(unicodeHost), hostsMapKey(asciiHost))
	tests.AssertEqual(t, hostsMapKey(asciiHost), strings.ToLower(asciiHost))
}

func TestSetFuncs(t *testing.T) {
	testErr := errors.New("test")
	marshalFunc := func(v any) ([]byte, error) {
		return nil, testErr
	}
	unmarshalFunc := func(data []byte, v any) error {
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
		return strings.Contains(contentType, "text")
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

func TestSetCommonQueryParamsFromValues(t *testing.T) {
	values := url.Values{}
	values.Add("test", "test")
	values.Add("key", "value")
	resp, err := tc().SetCommonQueryParamsFromValues(values).R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "key=value&test=test", resp.String())
}

func TestSetCommonQueryParamsFromStruct(t *testing.T) {
	type QueryParams struct {
		Test string `url:"test"`
		Key  string `url:"key"`
	}
	params := QueryParams{
		Test: "test",
		Key:  "value",
	}
	resp, err := tc().SetCommonQueryParamsFromStruct(params).R().Get("/query-parameter")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "key=value&test=test", resp.String())
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

	_, err = tc().SetRedirectPolicy(MaxRedirectPolicy(20)).SetRedirectPolicy(DefaultRedirectPolicy()).R().Get("/unlimited-redirect")
	tests.AssertNotNil(t, err)
	tests.AssertContains(t, err.Error(), "stopped after 10 redirects", true)

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
	}, []*http.Request{{
		Header: oldHeader,
	}})
	tests.AssertEqual(t, "test", newHeader.Get("Authorization"))
}

func TestSensitiveHeadersRedirectPolicy(t *testing.T) {
	// Cross-domain redirect: sensitive header should be stripped
	crossDomainReq := &http.Request{
		Header: http.Header{},
		URL:    &url.URL{Host: "evil.com"},
	}
	crossDomainReq.Header.Set("X-API-Key", "secret")
	via := []*http.Request{{
		Header: http.Header{},
		URL:    &url.URL{Host: "api.example.com"},
	}}
	via[0].Header.Set("X-API-Key", "secret")

	tc().SetRedirectPolicy(SensitiveHeadersRedirectPolicy("X-API-Key")).GetClient().CheckRedirect(crossDomainReq, via)
	tests.AssertEqual(t, "", crossDomainReq.Header.Get("X-API-Key"))

	// Same-domain redirect: sensitive header should be kept
	sameDomainReq := &http.Request{
		Header: http.Header{},
		URL:    &url.URL{Host: "sub.example.com"},
	}
	sameDomainReq.Header.Set("X-API-Key", "secret")
	viaSame := []*http.Request{{
		Header: http.Header{},
		URL:    &url.URL{Host: "api.example.com"},
	}}
	viaSame[0].Header.Set("X-API-Key", "secret")

	tc().SetRedirectPolicy(SensitiveHeadersRedirectPolicy("X-API-Key")).GetClient().CheckRedirect(sameDomainReq, viaSame)
	tests.AssertEqual(t, "secret", sameDomainReq.Header.Get("X-API-Key"))
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

func TestSetMultipartBoundaryFunc(t *testing.T) {
	delimiter := "test-delimiter"
	expectedContentType := fmt.Sprintf("multipart/form-data; boundary=%s", delimiter)
	resp, err := tc().
		SetMultipartBoundaryFunc(func() string {
			return delimiter
		}).R().
		EnableForceMultipart().
		SetFormData(
			map[string]string{
				"test": "test",
			}).
		Post("/content-type")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, expectedContentType, resp.String())
}

func TestFirefoxMultipartBoundaryFunc(t *testing.T) {
	r := regexp.MustCompile(`^-------------------------\d{1,10}\d{1,10}\d{1,10}$`)
	b := firefoxMultipartBoundaryFunc()
	tests.AssertEqual(t, true, r.MatchString(b))
}

func TestWebkitMultipartBoundaryFunc(t *testing.T) {
	r := regexp.MustCompile(`^----WebKitFormBoundary[0-9a-zA-Z]{16}$`)
	b := webkitMultipartBoundaryFunc()
	tests.AssertEqual(t, true, r.MatchString(b))
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

func TestSetTLSFingerprintSpec(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	server1 := httptest.NewTLSServer(http.HandlerFunc(handler))
	defer server1.Close()

	c := tc().SetTLSFingerprintSpec(func() utls.ClientHelloSpec {
		return utls.ClientHelloSpec{
			CipherSuites: []uint16{
				utls.TLS_AES_128_GCM_SHA256,
				utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			CompressionMethods: []byte{0x00},
			Extensions: []utls.TLSExtension{
				&utls.SNIExtension{},
				&utls.SupportedCurvesExtension{Curves: []utls.CurveID{
					utls.X25519,
					utls.CurveP256,
				}},
				&utls.SupportedPointsExtension{SupportedPoints: []byte{0x00}},
				&utls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{
					utls.ECDSAWithP256AndSHA256,
					utls.PSSWithSHA256,
					utls.PKCS1WithSHA256,
				}},
				&utls.KeyShareExtension{KeyShares: []utls.KeyShare{
					{Group: utls.X25519},
				}},
				&utls.SupportedVersionsExtension{Versions: []uint16{
					utls.VersionTLS13,
					utls.VersionTLS12,
				}},
			},
		}
	})
	if _, err := c.R().Get("/"); err != nil {
		t.Errorf("TestSetTLSFingerprintSpec failed on first handshake: %v", err)
	}
	if _, err := c.R().Get(server1.URL); err != nil {
		t.Errorf("TestSetTLSFingerprintSpec failed on consecutive handshake to different host: %v", err)
	}
}
