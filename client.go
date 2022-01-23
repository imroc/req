package req

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"github.com/imroc/req/v2/internal/util"
	"golang.org/x/net/publicsuffix"
	"io"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"os"
	"strings"
	"time"
)

var (
	hdrUserAgentKey   = "User-Agent"
	hdrUserAgentValue = "req/v2 (https://github.com/imroc/req)"
)

// DefaultClient returns the global default Client.
func DefaultClient() *Client {
	return defaultClient
}

// SetDefaultClient override the global default Client.
func SetDefaultClient(c *Client) {
	if c != nil {
		defaultClient = c
	}
}

var defaultClient *Client = C()

// Client is the req's http client.
type Client struct {
	HostURL       string
	PathParams    map[string]string
	QueryParams   urlpkg.Values
	Headers       http.Header
	Cookies       []*http.Cookie
	JSONMarshal   func(v interface{}) ([]byte, error)
	JSONUnmarshal func(data []byte, v interface{}) error
	XMLMarshal    func(v interface{}) ([]byte, error)
	XMLUnmarshal  func(data []byte, v interface{}) error
	Debug         bool

	disableAutoReadResponse bool
	scheme                  string
	log                     Logger
	t                       *Transport
	t2                      *http2Transport
	dumpOptions             *DumpOptions
	httpClient              *http.Client
	jsonDecoder             *json.Decoder
	beforeRequest           []RequestMiddleware
	udBeforeRequest         []RequestMiddleware
	afterResponse           []ResponseMiddleware
}

func cloneHeaders(hdrs http.Header) http.Header {
	if hdrs == nil {
		return nil
	}
	h := make(http.Header)
	for k, vs := range hdrs {
		for _, v := range vs {
			h.Add(k, v)
		}
	}
	return h
}

func cloneUrlValues(v urlpkg.Values) urlpkg.Values {
	if v == nil {
		return nil
	}
	vv := make(urlpkg.Values)
	for key, values := range v {
		for _, value := range values {
			vv.Add(key, value)
		}
	}
	return vv
}

func cloneMap(h map[string]string) map[string]string {
	if h == nil {
		return nil
	}
	m := make(map[string]string)
	for k, v := range h {
		m[k] = v
	}
	return m
}

// R create a new request.
func (c *Client) R() *Request {
	req := &http.Request{
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &Request{
		client:     c,
		RawRequest: req,
	}
}

func (c *Client) SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	if len(policies) == 0 {
		return c
	}
	c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		for _, f := range policies {
			if f == nil {
				continue
			}
			err := f(req, via)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return c
}

func (c *Client) DisableKeepAlives(disable bool) *Client {
	c.t.DisableKeepAlives = disable
	return c
}

func (c *Client) DisableCompression(disable bool) *Client {
	c.t.DisableCompression = disable
	return c
}

func (c *Client) SetTLSClientConfig(conf *tls.Config) *Client {
	c.t.TLSClientConfig = conf
	return c
}

func (c *Client) SetQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.SetQueryParam(k, v)
	}
	return c
}

func (c *Client) SetQueryParam(key, value string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	c.QueryParams.Set(key, value)
	return c
}

func (c *Client) SetQueryString(query string) *Client {
	params, err := urlpkg.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		if c.QueryParams == nil {
			c.QueryParams = make(urlpkg.Values)
		}
		for p, v := range params {
			for _, pv := range v {
				c.QueryParams.Add(p, pv)
			}
		}
	} else {
		c.log.Warnf("failed to parse query string (%s): %v", query, err)
	}
	return c
}

func (c *Client) SetCookie(hc *http.Cookie) *Client {
	c.Cookies = append(c.Cookies, hc)
	return c
}

func (c *Client) SetCookies(cs []*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cs...)
	return c
}

const (
	userAgentFirefox = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:95.0) Gecko/20100101 Firefox/95.0"
	userAgentChrome  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36"
)

func (c *Client) EnableDebug(enable bool) *Client {
	c.Debug = enable
	return c
}

// DevMode enables dump for requests and responses, and set user
// agent to pretend to be a web browser, Avoid returning abnormal
// data from some sites.
func (c *Client) DevMode() *Client {
	return c.EnableAutoDecodeTextType().
		EnableDumpAll().
		EnableDebug(true).
		SetUserAgent(userAgentChrome)
}

// SetScheme method sets custom scheme in the Resty client. It's way to override default.
// 		client.SetScheme("http")
func (c *Client) SetScheme(scheme string) *Client {
	if !util.IsStringEmpty(scheme) {
		c.scheme = strings.TrimSpace(scheme)
	}
	return c
}

// SetLogger set the logger for req, set to nil to disable logger.
func (c *Client) SetLogger(log Logger) *Client {
	if log == nil {
		c.log = &disableLogger{}
		return c
	}
	c.log = log
	return c
}

func (c *Client) GetResponseOptions() *ResponseOptions {
	if c.t.ResponseOptions == nil {
		c.t.ResponseOptions = &ResponseOptions{}
	}
	return c.t.ResponseOptions
}

// SetResponseOptions set the ResponseOptions for the underlying Transport.
func (c *Client) SetResponseOptions(opt *ResponseOptions) *Client {
	if opt == nil {
		return c
	}
	c.t.ResponseOptions = opt
	return c
}

// SetTimeout set the timeout for all requests.
func (c *Client) SetTimeout(d time.Duration) *Client {
	c.httpClient.Timeout = d
	return c
}

func (c *Client) GetDumpOptions() *DumpOptions {
	if c.dumpOptions == nil {
		c.dumpOptions = newDefaultDumpOptions()
	}
	return c.dumpOptions
}

func (c *Client) enableDump() {
	if c.t.dump != nil { // dump already started
		return
	}
	c.t.EnableDump(c.GetDumpOptions())
}

// EnableDumpToFile indicates that the content should dump to the specified filename.
func (c *Client) EnableDumpToFile(filename string) *Client {
	file, err := os.Create(filename)
	if err != nil {
		c.log.Errorf("create dump file error: %v", err)
		return c
	}
	c.GetDumpOptions().Output = file
	return c
}

// EnableDumpTo indicates that the content should dump to the specified destination.
func (c *Client) EnableDumpTo(output io.Writer) *Client {
	c.GetDumpOptions().Output = output
	c.enableDump()
	return c
}

// EnableDumpAsync indicates that the dump should be done asynchronously,
// can be used for debugging in production environment without
// affecting performance.
func (c *Client) EnableDumpAsync() *Client {
	o := c.GetDumpOptions()
	o.Async = true
	c.enableDump()
	return c
}

// EnableDumpOnlyResponse indicates that should dump the responses' head and response.
func (c *Client) EnableDumpOnlyResponse() *Client {
	o := c.GetDumpOptions()
	o.ResponseHeader = true
	o.ResponseBody = true
	o.RequestBody = false
	o.RequestHeader = false
	c.enableDump()
	return c
}

// EnableDumpOnlyRequest indicates that should dump the requests' head and response.
func (c *Client) EnableDumpOnlyRequest() *Client {
	o := c.GetDumpOptions()
	o.RequestHeader = true
	o.RequestBody = true
	o.ResponseBody = false
	o.ResponseHeader = false
	c.enableDump()
	return c
}

// EnableDumpOnlyBody indicates that should dump the body of requests and responses.
func (c *Client) EnableDumpOnlyBody() *Client {
	o := c.GetDumpOptions()
	o.RequestBody = true
	o.ResponseBody = true
	o.RequestHeader = false
	o.ResponseHeader = false
	c.enableDump()
	return c
}

// EnableDumpOnlyHeader indicates that should dump the head of requests and responses.
func (c *Client) EnableDumpOnlyHeader() *Client {
	o := c.GetDumpOptions()
	o.RequestHeader = true
	o.ResponseHeader = true
	o.RequestBody = false
	o.ResponseBody = false
	c.enableDump()
	return c
}

// EnableDumpAll indicates that should dump both requests and responses' head and body.
func (c *Client) EnableDumpAll() *Client {
	o := c.GetDumpOptions()
	o.RequestHeader = true
	o.RequestBody = true
	o.ResponseHeader = true
	o.ResponseBody = true
	c.enableDump()
	return c
}

// NewRequest is the alias of R()
func (c *Client) NewRequest() *Request {
	return c.R()
}

func (c *Client) DisableAutoReadResponse(disable bool) *Client {
	c.disableAutoReadResponse = disable
	return c
}

// EnableAutoDecodeAllType indicates that try autodetect and decode all content type.
func (c *Client) EnableAutoDecodeAllType() *Client {
	c.GetResponseOptions().AutoDecodeContentType = func(contentType string) bool {
		return true
	}
	return c
}

// EnableAutoDecodeTextType indicates that only try autodetect and decode the text content type.
func (c *Client) EnableAutoDecodeTextType() *Client {
	c.GetResponseOptions().AutoDecodeContentType = autoDecodeText
	return c
}

// SetUserAgent set the "User-Agent" header for all requests.
func (c *Client) SetUserAgent(userAgent string) *Client {
	return c.SetHeader(hdrUserAgentKey, userAgent)
}

func (c *Client) SetHeaders(hdrs map[string]string) *Client {
	for k, v := range hdrs {
		c.SetHeader(k, v)
	}
	return c
}

// SetHeader set the common header for all requests.
func (c *Client) SetHeader(key, value string) *Client {
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers.Set(key, value)
	return c
}

// EnableDump enables dump requests and responses,  allowing you
// to clearly see the content of all requests and responses，which
// is very convenient for debugging APIs.
func (c *Client) EnableDump(enable bool) *Client {
	if !enable {
		c.t.DisableDump()
		return c
	}
	c.enableDump()
	return c
}

// SetDumpOptions configures the underlying Transport's DumpOptions
func (c *Client) SetDumpOptions(opt *DumpOptions) *Client {
	if opt == nil {
		return c
	}
	c.dumpOptions = opt
	if c.t.dump != nil {
		c.t.dump.DumpOptions = opt
	}
	return c
}

// SetProxy set the proxy function.
func (c *Client) SetProxy(proxy func(*http.Request) (*urlpkg.URL, error)) *Client {
	c.t.Proxy = proxy
	return c
}

func (c *Client) SetProxyFromEnv() *Client {
	c.t.Proxy = http.ProxyFromEnvironment
	return c
}

func (c *Client) OnBeforeRequest(m RequestMiddleware) *Client {
	c.udBeforeRequest = append(c.udBeforeRequest, m)
	return c
}

func (c *Client) OnAfterResponse(m ResponseMiddleware) *Client {
	c.afterResponse = append(c.afterResponse, m)
	return c
}

func (c *Client) SetProxyURL(proxyUrl string) *Client {
	u, err := urlpkg.Parse(proxyUrl)
	if err != nil {
		c.log.Errorf("failed to parse proxy url %s: %v", proxyUrl, err)
		return c
	}
	c.t.Proxy = http.ProxyURL(u)
	return c
}

// NewClient is the alias of C
func NewClient() *Client {
	return C()
}

// Clone copy and returns the Client
func (c *Client) Clone() *Client {
	t := c.t.Clone()
	t2, _ := http2ConfigureTransports(t)
	cc := *c.httpClient
	cc.Transport = t
	return &Client{
		httpClient:              &cc,
		t:                       t,
		t2:                      t2,
		dumpOptions:             c.dumpOptions.Clone(),
		jsonDecoder:             c.jsonDecoder,
		Headers:                 cloneHeaders(c.Headers),
		PathParams:              cloneMap(c.PathParams),
		QueryParams:             cloneUrlValues(c.QueryParams),
		HostURL:                 c.HostURL,
		scheme:                  c.scheme,
		log:                     c.log,
		beforeRequest:           c.beforeRequest,
		udBeforeRequest:         c.udBeforeRequest,
		disableAutoReadResponse: c.disableAutoReadResponse,
	}
}

// C create a new client.
func C() *Client {
	t := &Transport{
		ForceAttemptHTTP2:     true,
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	t2, _ := http2ConfigureTransports(t)
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := &http.Client{
		Transport: t,
		Jar:       jar,
		Timeout:   2 * time.Minute,
	}
	beforeRequest := []RequestMiddleware{
		parseRequestURL,
		parseRequestHeader,
		parseRequestCookie,
	}
	afterResponse := []ResponseMiddleware{
		parseResponseBody,
		handleDownload,
	}
	c := &Client{
		beforeRequest: beforeRequest,
		afterResponse: afterResponse,
		log:           createLogger(),
		httpClient:    httpClient,
		t:             t,
		t2:            t2,
		JSONMarshal:   json.Marshal,
		JSONUnmarshal: json.Unmarshal,
		XMLMarshal:    xml.Marshal,
		XMLUnmarshal:  xml.Unmarshal,
	}
	return c
}

func setupRequest(r *Request) {
	setRequestURL(r.RawRequest, r.URL)
	setRequestHeaderAndCookie(r)
}

func (c *Client) do(r *Request) (resp *Response, err error) {

	for _, f := range r.client.udBeforeRequest {
		if err = f(r.client, r); err != nil {
			return
		}
	}

	for _, f := range r.client.beforeRequest {
		if err = f(r.client, r); err != nil {
			return
		}
	}

	setupRequest(r)

	if c.Debug {
		c.log.Debugf("%s %s", r.RawRequest.Method, r.RawRequest.URL.String())
	}

	httpResponse, err := c.httpClient.Do(r.RawRequest)
	if err != nil {
		return
	}

	resp = &Response{
		Request:  r,
		Response: httpResponse,
	}

	if !c.disableAutoReadResponse && !r.isSaveResponse { // auto read response body
		_, err = resp.Bytes()
		if err != nil {
			return
		}
	}

	for _, f := range r.client.afterResponse {
		if err = f(r.client, resp); err != nil {
			return
		}
	}
	return
}

func setRequestHeaderAndCookie(r *Request) {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.RawRequest.Header = r.Headers
	for _, cookie := range r.Cookies {
		r.RawRequest.AddCookie(cookie)
	}
}

func setRequestURL(r *http.Request, url string) error {
	// The host's colon:port should be normalized. See Issue 14836.
	u, err := urlpkg.Parse(url)
	if err != nil {
		return err
	}
	u.Host = removeEmptyPort(u.Host)
	r.URL = u
	r.Host = u.Host
	return nil
}
