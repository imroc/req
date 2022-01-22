package req

import (
	"encoding/json"
	"golang.org/x/net/publicsuffix"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
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
	log          Logger
	t            *Transport
	t2           *http2Transport
	dumpOptions  *DumpOptions
	httpClient   *http.Client
	jsonDecoder  *json.Decoder
	commonHeader map[string]string
}

func copyCommonHeader(h map[string]string) map[string]string {
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
		client:      c,
		httpRequest: req,
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
	return c.AutoDecodeTextType().
		Dump(true).
		SetLogger(NewLogger(os.Stdout)).
		UserAgent(userAgentChrome)
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

// ResponseOptions set the ResponseOptions for the underlying Transport.
func (c *Client) SetResponseOptions(opt *ResponseOptions) *Client {
	if opt == nil {
		return c
	}
	c.t.ResponseOptions = opt
	return c
}

// Timeout set the timeout for all requests.
func (c *Client) Timeout(d time.Duration) *Client {
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

// DumpToFile indicates that the content should dump to the specified filename.
func (c *Client) DumpToFile(filename string) *Client {
	file, err := os.Create(filename)
	if err != nil {
		logf(c.log, "create dump file error: %v", err)
		return c
	}
	c.GetDumpOptions().Output = file
	return c
}

// DumpTo indicates that the content should dump to the specified destination.
func (c *Client) DumpTo(output io.Writer) *Client {
	c.GetDumpOptions().Output = output
	c.enableDump()
	return c
}

// DumpAsync indicates that the dump should be done asynchronously,
// can be used for debugging in production environment without
// affecting performance.
func (c *Client) DumpAsync() *Client {
	o := c.GetDumpOptions()
	o.Async = true
	c.enableDump()
	return c
}

// DumpOnlyResponse indicates that should dump the responses' head and response.
func (c *Client) DumpOnlyResponse() *Client {
	o := c.GetDumpOptions()
	o.ResponseHead = true
	o.ResponseBody = true
	o.RequestBody = false
	o.RequestHead = false
	c.enableDump()
	return c
}

// DumpOnlyRequest indicates that should dump the requests' head and response.
func (c *Client) DumpOnlyRequest() *Client {
	o := c.GetDumpOptions()
	o.RequestHead = true
	o.RequestBody = true
	o.ResponseBody = false
	o.ResponseHead = false
	c.enableDump()
	return c
}

// DumpOnlyBody indicates that should dump the body of requests and responses.
func (c *Client) DumpOnlyBody() *Client {
	o := c.GetDumpOptions()
	o.RequestBody = true
	o.ResponseBody = true
	o.RequestHead = false
	o.ResponseHead = false
	c.enableDump()
	return c
}

// DumpOnlyHead indicates that should dump the head of requests and responses.
func (c *Client) DumpOnlyHead() *Client {
	o := c.GetDumpOptions()
	o.RequestHead = true
	o.ResponseHead = true
	o.RequestBody = false
	o.ResponseBody = false
	c.enableDump()
	return c
}

// DumpAll indicates that should dump both requests and responses' head and body.
func (c *Client) DumpAll() *Client {
	o := c.GetDumpOptions()
	o.RequestHead = true
	o.RequestBody = true
	o.ResponseHead = true
	o.ResponseBody = true
	c.enableDump()
	return c
}

// NewRequest is the alias of R()
func (c *Client) NewRequest() *Request {
	return c.R()
}

// AutoDecodeAllType indicates that try autodetect and decode all content type.
func (c *Client) AutoDecodeAllType() *Client {
	c.GetResponseOptions().AutoDecodeContentType = func(contentType string) bool {
		return true
	}
	return c
}

// AutoDecodeTextType indicates that only try autodetect and decode the text content type.
func (c *Client) AutoDecodeTextType() *Client {
	c.GetResponseOptions().AutoDecodeContentType = autoDecodeText
	return c
}

// UserAgent set the "User-Agent" header for all requests.
func (c *Client) UserAgent(userAgent string) *Client {
	return c.CommonHeader("User-Agent", userAgent)
}

// CommonHeader set the common header for all requests.
func (c *Client) CommonHeader(key, value string) *Client {
	if c.commonHeader == nil {
		c.commonHeader = make(map[string]string)
	}
	c.commonHeader[key] = value
	return c
}

// Dump if true, enables dump requests and responses,  allowing you
// to clearly see the content of all requests and responses，which
// is very convenient for debugging APIs.
// Dump if false, disable the dump behaviour.
func (c *Client) Dump(enable bool) *Client {
	if !enable {
		c.t.DisableDump()
		return c
	}
	c.enableDump()
	return c
}

// DumpOptions configures the underlying Transport's DumpOptions
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
		httpClient:   &cc,
		t:            t,
		t2:           t2,
		dumpOptions:  c.dumpOptions.Clone(),
		jsonDecoder:  c.jsonDecoder,
		commonHeader: copyCommonHeader(c.commonHeader),
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
	c := &Client{
		log:        &emptyLogger{},
		httpClient: httpClient,
		t:          t,
		t2:         t2,
	}
	return c
}
