package req

import (
	"encoding/json"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func DefaultClient() *Client {
	return defaultClient
}

func SetDefaultClient(c *Client) {
	if c != nil {
		defaultClient = c
	}
}

var defaultClient *Client = C()

type Client struct {
	t            *Transport
	t2           *http2Transport
	dumpOptions  *DumpOptions
	httpClient   *http.Client
	dialer       *net.Dialer
	jsonDecoder  *json.Decoder
	commonHeader map[string]string
}

func (c *Client) R() *Request {
	req := &http.Request{
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &Request{
		client:      c,
		httpRequest: req,
	}
}

func (c *Client) AutoDiscardResponseBody() *Client {
	return c.ResponseOptions(DiscardResponseBody())
}

// TestMode is like DebugMode, but discard response body, so you can
// dump responses without read response body
func (c *Client) TestMode() *Client {
	return c.DebugMode().AutoDiscardResponseBody()
}

// DebugMode enables dump for requests and responses, and set user
// agent to pretend to be a web browser, Avoid returning abnormal
// data from some sites.
func (c *Client) DebugMode() *Client {
	return c.AutoDecodeTextContent().EnableDump(DumpAll()).UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36")
}

func (c *Client) ResponseOptions(opts ...ResponseOption) *Client {
	for _, opt := range opts {
		opt(&c.t.ResponseOptions)
	}
	return c
}

func (c *Client) Timeout(d time.Duration) *Client {
	c.httpClient.Timeout = d
	return c
}

// NewRequest is the alias of R()
func (c *Client) NewRequest() *Request {
	return c.R()
}

func (c *Client) DisableDump() *Client {
	c.t.DisableDump()
	return c
}

func (c *Client) AutoDecodeTextContent() *Client {
	return c.ResponseOptions(AutoDecodeTextContent())
}

func (c *Client) UserAgent(userAgent string) *Client {
	return c.CommonHeader("User-Agent", userAgent)
}

func (c *Client) CommonHeader(key, value string) *Client {
	if c.commonHeader == nil {
		c.commonHeader = make(map[string]string)
	}
	c.commonHeader[key] = value
	return c
}

// EnableDump enables dump requests and responses,  allowing you
// to clearly see the content of all requests and responses，which
// is very convenient for debugging APIs.
// EnableDump accepet options for custom the dump behavior:
// 1. DumpAsync: dump asynchronously, can be used for debugging in
//  production environment without affecting performance.
// 2. DumpHead, DumpBody,
func (c *Client) EnableDump(opts ...DumpOption) *Client {
	if len(opts) > 0 {
		if c.dumpOptions == nil {
			c.dumpOptions = &DumpOptions{}
		}
		c.dumpOptions.set(opts...)
	}
	c.t.EnableDump(c.dumpOptions)
	return c
}

// NewClient is the alias of C()
func NewClient() *Client {
	return C()
}

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
	c := &Client{
		httpClient: httpClient,
		t:          t,
		t2:         t2,
	}
	return c
}
