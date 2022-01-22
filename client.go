package req

import (
	"encoding/json"
	"github.com/imroc/req/v2/internal/util"
	"golang.org/x/net/publicsuffix"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
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
	HostURL         string
	PathParams      map[string]string
	QueryParams     url.Values
	scheme          string
	log             Logger
	t               *Transport
	t2              *http2Transport
	dumpOptions     *DumpOptions
	httpClient      *http.Client
	jsonDecoder     *json.Decoder
	Headers         map[string]string
	beforeRequest   []RequestMiddleware
	udBeforeRequest []RequestMiddleware
}

func cloneUrlValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}
	vv := make(url.Values)
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

func (c *Client) AutoDiscardResponseBody() *Client {
	c.GetResponseOptions().AutoDiscard = true
	return c
}

// TestMode is like DebugMode, but discard response body, so you can
// dump responses without read response body
func (c *Client) TestMode() *Client {
	return c.DebugMode().AutoDiscardResponseBody()
}

const (
	userAgentFirefox = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:95.0) Gecko/20100101 Firefox/95.0"
	userAgentChrome  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36"
)

// DebugMode enables dump for requests and responses, and set user
// agent to pretend to be a web browser, Avoid returning abnormal
// data from some sites.
func (c *Client) DebugMode() *Client {
	return c.EnableAutoDecodeTextType().
		EnableDumpAll().
		SetLogger(NewLogger(os.Stdout)).
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

// SetLogger set the logger for req.
func (c *Client) SetLogger(log Logger) *Client {
	if log == nil {
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
		logf(c.log, "create dump file error: %v", err)
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
	return c.SetHeader("User-Agent", userAgent)
}

// SetHeader set the common header for all requests.
func (c *Client) SetHeader(key, value string) *Client {
	if c.Headers == nil {
		c.Headers = make(map[string]string)
	}
	c.Headers[key] = value
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
func (c *Client) SetProxy(proxy func(*http.Request) (*url.URL, error)) *Client {
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

func (c *Client) SetProxyURL(proxyUrl string) *Client {
	u, err := url.Parse(proxyUrl)
	if err != nil {
		logf(c.log, "failed to parse proxy url %s: %v", proxyUrl, err)
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
		httpClient:      &cc,
		t:               t,
		t2:              t2,
		dumpOptions:     c.dumpOptions.Clone(),
		jsonDecoder:     c.jsonDecoder,
		Headers:         cloneMap(c.Headers),
		PathParams:      cloneMap(c.PathParams),
		QueryParams:     cloneUrlValues(c.QueryParams),
		HostURL:         c.HostURL,
		scheme:          c.scheme,
		log:             c.log,
		beforeRequest:   c.beforeRequest,
		udBeforeRequest: c.udBeforeRequest,
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
	}
	c := &Client{
		beforeRequest: beforeRequest,
		log:           &emptyLogger{},
		httpClient:    httpClient,
		t:             t,
		t2:            t2,
	}
	return c
}
