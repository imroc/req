package req

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/imroc/req/v3/internal/util"
	"golang.org/x/net/publicsuffix"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
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
	BaseURL               string
	PathParams            map[string]string
	QueryParams           urlpkg.Values
	Headers               http.Header
	Cookies               []*http.Cookie
	FormData              urlpkg.Values
	DebugLog              bool
	AllowGetMethodPayload bool

	jsonMarshal             func(v interface{}) ([]byte, error)
	jsonUnmarshal           func(data []byte, v interface{}) error
	xmlMarshal              func(v interface{}) ([]byte, error)
	xmlUnmarshal            func(data []byte, v interface{}) error
	trace                   bool
	outputDirectory         string
	disableAutoReadResponse bool
	scheme                  string
	log                     Logger
	t                       *Transport
	t2                      *http2Transport
	dumpOptions             *DumpOptions
	httpClient              *http.Client
	beforeRequest           []RequestMiddleware
	udBeforeRequest         []RequestMiddleware
	afterResponse           []ResponseMiddleware
}

// R is a global wrapper methods which delegated
// to the default client's R().
func R() *Request {
	return defaultClient.R()
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

// SetCommonFormDataFromValues is a global wrapper methods which delegated
// to the default client's SetCommonFormDataFromValues.
func SetCommonFormDataFromValues(data urlpkg.Values) *Client {
	return defaultClient.SetCommonFormDataFromValues(data)
}

// SetCommonFormDataFromValues set the form data from url.Values for all requests
// which request method allows payload.
func (c *Client) SetCommonFormDataFromValues(data urlpkg.Values) *Client {
	if c.FormData == nil {
		c.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		for _, kv := range v {
			c.FormData.Add(k, kv)
		}
	}
	return c
}

// SetCommonFormData is a global wrapper methods which delegated
// to the default client's SetCommonFormData.
func SetCommonFormData(data map[string]string) *Client {
	return defaultClient.SetCommonFormData(data)
}

// SetCommonFormData set the form data from map for all requests
// which request method allows payload.
func (c *Client) SetCommonFormData(data map[string]string) *Client {
	if c.FormData == nil {
		c.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		c.FormData.Set(k, v)
	}
	return c
}

// SetBaseURL is a global wrapper methods which delegated
// to the default client's SetBaseURL.
func SetBaseURL(u string) *Client {
	return defaultClient.SetBaseURL(u)
}

// SetBaseURL set the default base URL, will be used if request URL is
// a relative URL.
func (c *Client) SetBaseURL(u string) *Client {
	c.BaseURL = strings.TrimRight(u, "/")
	return c
}

// SetOutputDirectory is a global wrapper methods which delegated
// to the default client's SetOutputDirectory.
func SetOutputDirectory(dir string) *Client {
	return defaultClient.SetOutputDirectory(dir)
}

// SetOutputDirectory set output directory that response will
// be downloaded to.
func (c *Client) SetOutputDirectory(dir string) *Client {
	c.outputDirectory = dir
	return c
}

// SetCertFromFile is a global wrapper methods which delegated
// to the default client's SetCertFromFile.
func SetCertFromFile(certFile, keyFile string) *Client {
	return defaultClient.SetCertFromFile(certFile, keyFile)
}

// SetCertFromFile helps to set client certificates from cert and key file.
func (c *Client) SetCertFromFile(certFile, keyFile string) *Client {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		c.log.Errorf("failed to load client cert: %v", err)
		return c
	}
	config := c.GetTLSClientConfig()
	config.Certificates = append(config.Certificates, cert)
	return c
}

// SetCerts is a global wrapper methods which delegated
// to the default client's SetCerts.
func SetCerts(certs ...tls.Certificate) *Client {
	return defaultClient.SetCerts(certs...)
}

// SetCerts set client certificates.
func (c *Client) SetCerts(certs ...tls.Certificate) *Client {
	config := c.GetTLSClientConfig()
	config.Certificates = append(config.Certificates, certs...)
	return c
}

func (c *Client) appendRootCertData(data []byte) {
	config := c.GetTLSClientConfig()
	if config.RootCAs == nil {
		config.RootCAs = x509.NewCertPool()
	}
	config.RootCAs.AppendCertsFromPEM(data)
	return
}

// SetRootCertFromString is a global wrapper methods which delegated
// to the default client's SetRootCertFromString.
func SetRootCertFromString(pemContent string) *Client {
	return defaultClient.SetRootCertFromString(pemContent)
}

// SetRootCertFromString set root certificates from string.
func (c *Client) SetRootCertFromString(pemContent string) *Client {
	c.appendRootCertData([]byte(pemContent))
	return c
}

// SetRootCertsFromFile is a global wrapper methods which delegated
// to the default client's SetRootCertsFromFile.
func SetRootCertsFromFile(pemFiles ...string) *Client {
	return defaultClient.SetRootCertsFromFile(pemFiles...)
}

// SetRootCertsFromFile set root certificates from files.
func (c *Client) SetRootCertsFromFile(pemFiles ...string) *Client {
	for _, pemFile := range pemFiles {
		rootPemData, err := ioutil.ReadFile(pemFile)
		if err != nil {
			c.log.Errorf("failed to read root cert file: %v", err)
			return c
		}
		c.appendRootCertData(rootPemData)
	}
	return c
}

// GetTLSClientConfig is a global wrapper methods which delegated
// to the default client's GetTLSClientConfig.
func GetTLSClientConfig() *tls.Config {
	return defaultClient.GetTLSClientConfig()
}

// GetTLSClientConfig return the underlying tls.Config.
func (c *Client) GetTLSClientConfig() *tls.Config {
	if c.t.TLSClientConfig == nil {
		c.t.TLSClientConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		}
	}
	return c.t.TLSClientConfig
}

func (c *Client) defaultCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if c.DebugLog {
		c.log.Debugf("<redirect> %s %s", req.Method, req.URL.String())
	}
	return nil
}

// SetRedirectPolicy is a global wrapper methods which delegated
// to the default client's SetRedirectPolicy.
func SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	return defaultClient.SetRedirectPolicy(policies...)
}

// SetRedirectPolicy set the RedirectPolicy which controls the behavior of receiving redirect
// responses (usually responses with 301 and 302 status code), see the predefined
// AllowedDomainRedirectPolicy, AllowedHostRedirectPolicy, MaxRedirectPolicy, NoRedirectPolicy,
// SameDomainRedirectPolicy and SameHostRedirectPolicy.
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
		if c.DebugLog {
			c.log.Debugf("<redirect> %s %s", req.Method, req.URL.String())
		}
		return nil
	}
	return c
}

// DisableKeepAlives is a global wrapper methods which delegated
// to the default client's DisableKeepAlives.
func DisableKeepAlives() *Client {
	return defaultClient.DisableKeepAlives()
}

// DisableKeepAlives disable the HTTP keep-alives (enabled by default)
// and will only use the connection to the server for a single
// HTTP request.
//
// This is unrelated to the similarly named TCP keep-alives.
func (c *Client) DisableKeepAlives() *Client {
	c.t.DisableKeepAlives = true
	return c
}

// EnableKeepAlives is a global wrapper methods which delegated
// to the default client's EnableKeepAlives.
func EnableKeepAlives() *Client {
	return defaultClient.EnableKeepAlives()
}

// EnableKeepAlives enables HTTP keep-alives (enabled by default).
func (c *Client) EnableKeepAlives() *Client {
	c.t.DisableKeepAlives = false
	return c
}

// DisableCompression is a global wrapper methods which delegated
// to the default client's DisableCompression.
func DisableCompression() *Client {
	return defaultClient.DisableCompression()
}

// DisableCompression disables the compression (enabled by default),
// which prevents the Transport from requesting compression
// with an "Accept-Encoding: gzip" request header when the
// Request contains no existing Accept-Encoding value. If
// the Transport requests gzip on its own and gets a gzipped
// response, it's transparently decoded in the Response.Body.
// However, if the user explicitly requested gzip it is not
// automatically uncompressed.
func (c *Client) DisableCompression() *Client {
	c.t.DisableCompression = true
	return c
}

// EnableCompression is a global wrapper methods which delegated
// to the default client's EnableCompression.
func EnableCompression() *Client {
	return defaultClient.EnableCompression()
}

// EnableCompression enables the compression (enabled by default).
func (c *Client) EnableCompression() *Client {
	c.t.DisableCompression = false
	return c
}

// SetTLSClientConfig is a global wrapper methods which delegated
// to the default client's SetTLSClientConfig.
func SetTLSClientConfig(conf *tls.Config) *Client {
	return defaultClient.SetTLSClientConfig(conf)
}

// SetTLSClientConfig set the TLS client config. Be careful! Usually
// you don't need this, you can directly set the tls configuration with
// methods like EnableInsecureSkipVerify, SetCerts etc. Or you can call
// GetTLSClientConfig to get the current tls configuration to avoid
// overwriting some important configurations, such as not setting NextProtos
// will not use http2 by default.
func (c *Client) SetTLSClientConfig(conf *tls.Config) *Client {
	c.t.TLSClientConfig = conf
	return c
}

// EnableInsecureSkipVerify is a global wrapper methods which delegated
// to the default client's EnableInsecureSkipVerify.
func EnableInsecureSkipVerify() *Client {
	return defaultClient.EnableInsecureSkipVerify()
}

// EnableInsecureSkipVerify enable send https without verifing
// the server's certificates (disabled by default).
func (c *Client) EnableInsecureSkipVerify() *Client {
	c.GetTLSClientConfig().InsecureSkipVerify = true
	return c
}

// DisableInsecureSkipVerify is a global wrapper methods which delegated
// to the default client's DisableInsecureSkipVerify.
func DisableInsecureSkipVerify() *Client {
	return defaultClient.DisableInsecureSkipVerify()
}

// DisableInsecureSkipVerify disable send https without verifing
// the server's certificates (disabled by default).
func (c *Client) DisableInsecureSkipVerify() *Client {
	c.GetTLSClientConfig().InsecureSkipVerify = false
	return c
}

// SetCommonQueryParams is a global wrapper methods which delegated
// to the default client's SetCommonQueryParams.
func SetCommonQueryParams(params map[string]string) *Client {
	return defaultClient.SetCommonQueryParams(params)
}

// SetCommonQueryParams set URL query parameters with a map
// for all requests.
func (c *Client) SetCommonQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.SetCommonQueryParam(k, v)
	}
	return c
}

// AddCommonQueryParam is a global wrapper methods which delegated
// to the default client's AddCommonQueryParam.
func AddCommonQueryParam(key, value string) *Client {
	return defaultClient.AddCommonQueryParam(key, value)
}

// AddCommonQueryParam add a URL query parameter with a key-value
// pair for all requests.
func (c *Client) AddCommonQueryParam(key, value string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	c.QueryParams.Add(key, value)
	return c
}

// SetCommonQueryParam is a global wrapper methods which delegated
// to the default client's SetCommonQueryParam.
func SetCommonQueryParam(key, value string) *Client {
	return defaultClient.SetCommonQueryParam(key, value)
}

// SetCommonQueryParam set a URL query parameter with a key-value
// pair for all requests.
func (c *Client) SetCommonQueryParam(key, value string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	c.QueryParams.Set(key, value)
	return c
}

// SetCommonQueryString is a global wrapper methods which delegated
// to the default client's SetCommonQueryString.
func SetCommonQueryString(query string) *Client {
	return defaultClient.SetCommonQueryString(query)
}

// SetCommonQueryString set URL query parameters with a raw query string
// for all requests.
func (c *Client) SetCommonQueryString(query string) *Client {
	params, err := urlpkg.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		c.log.Warnf("failed to parse query string (%s): %v", query, err)
		return c
	}
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	for p, v := range params {
		for _, pv := range v {
			c.QueryParams.Add(p, pv)
		}
	}
	return c
}

// SetCommonCookies is a global wrapper methods which delegated
// to the default client's SetCommonCookies.
func SetCommonCookies(cookies ...*http.Cookie) *Client {
	return defaultClient.SetCommonCookies(cookies...)
}

// SetCommonCookies set HTTP cookies for all requests.
func (c *Client) SetCommonCookies(cookies ...*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cookies...)
	return c
}

// DisableDebugLog is a global wrapper methods which delegated
// to the default client's DisableDebugLog.
func DisableDebugLog() *Client {
	return defaultClient.DisableDebugLog()
}

// DisableDebugLog disable debug level log (disabled by default).
func (c *Client) DisableDebugLog() *Client {
	c.DebugLog = false
	return c
}

// EnableDebugLog is a global wrapper methods which delegated
// to the default client's EnableDebugLog.
func EnableDebugLog() *Client {
	return defaultClient.EnableDebugLog()
}

// EnableDebugLog enable debug level log (disabled by default).
func (c *Client) EnableDebugLog() *Client {
	c.DebugLog = true
	return c
}

// DevMode is a global wrapper methods which delegated
// to the default client's DevMode.
func DevMode() *Client {
	return defaultClient.DevMode()
}

// DevMode enables:
// 1. Dump content of all requests and responses to see details.
// 2. Output debug level log for deeper insights.
// 3. Trace all requests, so you can get trace info to analyze performance.
// 4. Set User-Agent to pretend to be a web browser, avoid returning abnormal data from some sites.
func (c *Client) DevMode() *Client {
	return c.EnableDumpAll().
		EnableDebugLog().
		EnableTraceAll().
		SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36")
}

// SetScheme is a global wrapper methods which delegated
// to the default client's SetScheme.
func SetScheme(scheme string) *Client {
	return defaultClient.SetScheme(scheme)
}

// SetScheme set the default scheme for client, will be used when
// there is no scheme in the request URL (e.g. "github.com/imroc/req").
func (c *Client) SetScheme(scheme string) *Client {
	if !util.IsStringEmpty(scheme) {
		c.scheme = strings.TrimSpace(scheme)
	}
	return c
}

// SetLogger is a global wrapper methods which delegated
// to the default client's SetLogger.
func SetLogger(log Logger) *Client {
	return defaultClient.SetLogger(log)
}

// SetLogger set the customized logger for client, will disable log if set to nil.
func (c *Client) SetLogger(log Logger) *Client {
	if log == nil {
		c.log = &disableLogger{}
		return c
	}
	c.log = log
	return c
}

func (c *Client) getResponseOptions() *ResponseOptions {
	if c.t.ResponseOptions == nil {
		c.t.ResponseOptions = &ResponseOptions{}
	}
	return c.t.ResponseOptions
}

// SetTimeout is a global wrapper methods which delegated
// to the default client's SetTimeout.
func SetTimeout(d time.Duration) *Client {
	return defaultClient.SetTimeout(d)
}

// SetTimeout set timeout for all requests.
func (c *Client) SetTimeout(d time.Duration) *Client {
	c.httpClient.Timeout = d
	return c
}

func (c *Client) getDumpOptions() *DumpOptions {
	if c.dumpOptions == nil {
		c.dumpOptions = newDefaultDumpOptions()
	}
	return c.dumpOptions
}

// EnableDumpAll is a global wrapper methods which delegated
// to the default client's EnableDumpAll.
func EnableDumpAll() *Client {
	return defaultClient.EnableDumpAll()
}

// EnableDumpAll enable dump for all requests, including
// all content for the request and response by default.
func (c *Client) EnableDumpAll() *Client {
	if c.t.dump != nil { // dump already started
		return c
	}
	c.t.EnableDump(c.getDumpOptions())
	return c
}

// EnableDumpAllToFile is a global wrapper methods which delegated
// to the default client's EnableDumpAllToFile.
func EnableDumpAllToFile(filename string) *Client {
	return defaultClient.EnableDumpAllToFile(filename)
}

// EnableDumpAllToFile enable dump for all requests and output
// to the specified file.
func (c *Client) EnableDumpAllToFile(filename string) *Client {
	file, err := os.Create(filename)
	if err != nil {
		c.log.Errorf("create dump file error: %v", err)
		return c
	}
	c.getDumpOptions().Output = file
	c.EnableDumpAll()
	return c
}

// EnableDumpAllTo is a global wrapper methods which delegated
// to the default client's EnableDumpAllTo.
func EnableDumpAllTo(output io.Writer) *Client {
	return defaultClient.EnableDumpAllTo(output)
}

// EnableDumpAllTo enable dump for all requests and output to
// the specified io.Writer.
func (c *Client) EnableDumpAllTo(output io.Writer) *Client {
	c.getDumpOptions().Output = output
	c.EnableDumpAll()
	return c
}

// EnableDumpAllAsync is a global wrapper methods which delegated
// to the default client's EnableDumpAllAsync.
func EnableDumpAllAsync() *Client {
	return defaultClient.EnableDumpAllAsync()
}

// EnableDumpAllAsync enable dump for all requests and output
// asynchronously, can be used for debugging in production
// environment without affecting performance.
func (c *Client) EnableDumpAllAsync() *Client {
	o := c.getDumpOptions()
	o.Async = true
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutRequestBody is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutRequestBody.
func EnableDumpAllWithoutRequestBody() *Client {
	return defaultClient.EnableDumpAllWithoutRequestBody()
}

// EnableDumpAllWithoutRequestBody enable dump for all requests without
// request body, can be used in the upload request to avoid dumping the
// unreadable binary content.
func (c *Client) EnableDumpAllWithoutRequestBody() *Client {
	o := c.getDumpOptions()
	o.RequestBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutResponseBody is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutResponseBody.
func EnableDumpAllWithoutResponseBody() *Client {
	return defaultClient.EnableDumpAllWithoutResponseBody()
}

// EnableDumpAllWithoutResponseBody enable dump for all requests without
// response body, can be used in the download request to avoid dumping the
// unreadable binary content.
func (c *Client) EnableDumpAllWithoutResponseBody() *Client {
	o := c.getDumpOptions()
	o.ResponseBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutResponse is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutResponse.
func EnableDumpAllWithoutResponse() *Client {
	return defaultClient.EnableDumpAllWithoutResponse()
}

// EnableDumpAllWithoutResponse enable dump for all requests without response,
// can be used if you only care about the request.
func (c *Client) EnableDumpAllWithoutResponse() *Client {
	o := c.getDumpOptions()
	o.ResponseBody = false
	o.ResponseHeader = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutRequest is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutRequest.
func EnableDumpAllWithoutRequest() *Client {
	return defaultClient.EnableDumpAllWithoutRequest()
}

// EnableDumpAllWithoutRequest enables dump for all requests without request,
// can be used if you only care about the response.
func (c *Client) EnableDumpAllWithoutRequest() *Client {
	o := c.getDumpOptions()
	o.RequestHeader = false
	o.RequestBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutHeader is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutHeader.
func EnableDumpAllWithoutHeader() *Client {
	return defaultClient.EnableDumpAllWithoutHeader()
}

// EnableDumpAllWithoutHeader enable dump for all requests without header,
// can be used if you only care about the body.
func (c *Client) EnableDumpAllWithoutHeader() *Client {
	o := c.getDumpOptions()
	o.RequestHeader = false
	o.ResponseHeader = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutBody is a global wrapper methods which delegated
// to the default client's EnableDumpAllWithoutBody.
func EnableDumpAllWithoutBody() *Client {
	return defaultClient.EnableDumpAllWithoutBody()
}

// EnableDumpAllWithoutBody enable dump for all requests without body,
// can be used if you only care about the header.
func (c *Client) EnableDumpAllWithoutBody() *Client {
	o := c.getDumpOptions()
	o.RequestBody = false
	o.ResponseBody = false
	c.EnableDumpAll()
	return c
}

// NewRequest is a global wrapper methods which delegated
// to the default client's NewRequest.
func NewRequest() *Request {
	return defaultClient.R()
}

// NewRequest is the alias of R()
func (c *Client) NewRequest() *Request {
	return c.R()
}

// DisableAutoReadResponse is a global wrapper methods which delegated
// to the default client's DisableAutoReadResponse.
func DisableAutoReadResponse() *Client {
	return defaultClient.DisableAutoReadResponse()
}

// DisableAutoReadResponse disable read response body automatically (enabled by default).
func (c *Client) DisableAutoReadResponse() *Client {
	c.disableAutoReadResponse = true
	return c
}

// EnableAutoReadResponse is a global wrapper methods which delegated
// to the default client's EnableAutoReadResponse.
func EnableAutoReadResponse() *Client {
	return defaultClient.EnableAutoReadResponse()
}

// EnableAutoReadResponse enable read response body automatically (enabled by default).
func (c *Client) EnableAutoReadResponse() *Client {
	c.disableAutoReadResponse = false
	return c
}

// SetAutoDecodeContentType is a global wrapper methods which delegated
// to the default client's SetAutoDecodeContentType.
func SetAutoDecodeContentType(contentTypes ...string) *Client {
	return defaultClient.SetAutoDecodeContentType(contentTypes...)
}

// SetAutoDecodeContentType set the content types that will be auto-detected and decode
// to utf-8 (e.g. "json", "xml", "html", "text").
func (c *Client) SetAutoDecodeContentType(contentTypes ...string) *Client {
	opt := c.getResponseOptions()
	opt.AutoDecodeContentType = autoDecodeContentTypeFunc(contentTypes...)
	return c
}

// SetAutoDecodeContentTypeFunc is a global wrapper methods which delegated
// to the default client's SetAutoDecodeAllTypeFunc.
func SetAutoDecodeContentTypeFunc(fn func(contentType string) bool) *Client {
	return defaultClient.SetAutoDecodeContentTypeFunc(fn)
}

// SetAutoDecodeContentTypeFunc set the function that determines whether the
// specified `Content-Type` should be auto-detected and decode to utf-8.
func (c *Client) SetAutoDecodeContentTypeFunc(fn func(contentType string) bool) *Client {
	opt := c.getResponseOptions()
	opt.AutoDecodeContentType = fn
	return c
}

// SetAutoDecodeAllContentType is a global wrapper methods which delegated
// to the default client's SetAutoDecodeAllContentType.
func SetAutoDecodeAllContentType() *Client {
	return defaultClient.SetAutoDecodeAllContentType()
}

// SetAutoDecodeAllContentType enable try auto-detect charset and decode all
// content type to utf-8.
func (c *Client) SetAutoDecodeAllContentType() *Client {
	opt := c.getResponseOptions()
	opt.AutoDecodeContentType = func(contentType string) bool {
		return true
	}
	return c
}

// DisableAutoDecode is a global wrapper methods which delegated
// to the default client's DisableAutoDecode.
func DisableAutoDecode() *Client {
	return defaultClient.DisableAutoDecode()
}

// DisableAutoDecode disable auto-detect charset and decode to utf-8
// (enabled by default).
func (c *Client) DisableAutoDecode() *Client {
	c.getResponseOptions().DisableAutoDecode = true
	return c
}

// EnableAutoDecode is a global wrapper methods which delegated
// to the default client's EnableAutoDecode.
func EnableAutoDecode() *Client {
	return defaultClient.EnableAutoDecode()
}

// EnableAutoDecode enable auto-detect charset and decode to utf-8
// (enabled by default).
func (c *Client) EnableAutoDecode() *Client {
	c.getResponseOptions().DisableAutoDecode = true
	return c
}

// SetUserAgent is a global wrapper methods which delegated
// to the default client's SetUserAgent.
func SetUserAgent(userAgent string) *Client {
	return defaultClient.SetUserAgent(userAgent)
}

// SetUserAgent set the "User-Agent" header for all requests.
func (c *Client) SetUserAgent(userAgent string) *Client {
	return c.SetCommonHeader(hdrUserAgentKey, userAgent)
}

// SetCommonBearerAuthToken is a global wrapper methods which delegated
// to the default client's SetCommonBearerAuthToken.
func SetCommonBearerAuthToken(token string) *Client {
	return defaultClient.SetCommonBearerAuthToken(token)
}

// SetCommonBearerAuthToken set the bearer auth token for all requests.
func (c *Client) SetCommonBearerAuthToken(token string) *Client {
	return c.SetCommonHeader("Authorization", "Bearer "+token)
}

// SetCommonBasicAuth is a global wrapper methods which delegated
// to the default client's SetCommonBasicAuth.
func SetCommonBasicAuth(username, password string) *Client {
	return defaultClient.SetCommonBasicAuth(username, password)
}

// SetCommonBasicAuth set the basic auth for all requests.
func (c *Client) SetCommonBasicAuth(username, password string) *Client {
	c.SetCommonHeader("Authorization", util.BasicAuthHeaderValue(username, password))
	return c
}

// SetCommonHeaders is a global wrapper methods which delegated
// to the default client's SetCommonHeaders.
func SetCommonHeaders(hdrs map[string]string) *Client {
	return defaultClient.SetCommonHeaders(hdrs)
}

// SetCommonHeaders set headers for all requests.
func (c *Client) SetCommonHeaders(hdrs map[string]string) *Client {
	for k, v := range hdrs {
		c.SetCommonHeader(k, v)
	}
	return c
}

// SetCommonHeader is a global wrapper methods which delegated
// to the default client's SetCommonHeader.
func SetCommonHeader(key, value string) *Client {
	return defaultClient.SetCommonHeader(key, value)
}

// SetCommonHeader set a header for all requests.
func (c *Client) SetCommonHeader(key, value string) *Client {
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers.Set(key, value)
	return c
}

// SetCommonContentType is a global wrapper methods which delegated
// to the default client's SetCommonContentType.
func SetCommonContentType(ct string) *Client {
	return defaultClient.SetCommonContentType(ct)
}

// SetCommonContentType set the `Content-Type` header for all requests.
func (c *Client) SetCommonContentType(ct string) *Client {
	c.SetCommonHeader(hdrContentTypeKey, ct)
	return c
}

// DisableDumpAll is a global wrapper methods which delegated
// to the default client's DisableDumpAll.
func DisableDumpAll() *Client {
	return defaultClient.DisableDumpAll()
}

// DisableDumpAll disable dump for all requests.
func (c *Client) DisableDumpAll() *Client {
	c.t.DisableDump()
	return c
}

// SetCommonDumpOptions is a global wrapper methods which delegated
// to the default client's SetCommonDumpOptions.
func SetCommonDumpOptions(opt *DumpOptions) *Client {
	return defaultClient.SetCommonDumpOptions(opt)
}

// SetCommonDumpOptions configures the underlying Transport's DumpOptions
// for all requests.
func (c *Client) SetCommonDumpOptions(opt *DumpOptions) *Client {
	if opt == nil {
		return c
	}
	c.dumpOptions = opt
	if c.t.dump != nil {
		c.t.dump.DumpOptions = opt
	}
	return c
}

// SetProxy is a global wrapper methods which delegated
// to the default client's SetProxy.
func SetProxy(proxy func(*http.Request) (*urlpkg.URL, error)) *Client {
	return defaultClient.SetProxy(proxy)
}

// SetProxy set the proxy function.
func (c *Client) SetProxy(proxy func(*http.Request) (*urlpkg.URL, error)) *Client {
	c.t.Proxy = proxy
	return c
}

// OnBeforeRequest is a global wrapper methods which delegated
// to the default client's OnBeforeRequest.
func OnBeforeRequest(m RequestMiddleware) *Client {
	return defaultClient.OnBeforeRequest(m)
}

// OnBeforeRequest add a request middleware which hooks before request sent.
func (c *Client) OnBeforeRequest(m RequestMiddleware) *Client {
	c.udBeforeRequest = append(c.udBeforeRequest, m)
	return c
}

// OnAfterResponse is a global wrapper methods which delegated
// to the default client's OnAfterResponse.
func OnAfterResponse(m ResponseMiddleware) *Client {
	return defaultClient.OnAfterResponse(m)
}

// OnAfterResponse add a response middleware which hooks after response received.
func (c *Client) OnAfterResponse(m ResponseMiddleware) *Client {
	c.afterResponse = append(c.afterResponse, m)
	return c
}

// SetProxyURL is a global wrapper methods which delegated
// to the default client's SetProxyURL.
func SetProxyURL(proxyUrl string) *Client {
	return defaultClient.SetProxyURL(proxyUrl)
}

// SetProxyURL set proxy from the proxy URL.
func (c *Client) SetProxyURL(proxyUrl string) *Client {
	u, err := urlpkg.Parse(proxyUrl)
	if err != nil {
		c.log.Errorf("failed to parse proxy url %s: %v", proxyUrl, err)
		return c
	}
	c.t.Proxy = http.ProxyURL(u)
	return c
}

// DisableTraceAll is a global wrapper methods which delegated
// to the default client's DisableTraceAll.
func DisableTraceAll() *Client {
	return defaultClient.DisableTraceAll()
}

// DisableTraceAll disable trace for all requests.
func (c *Client) DisableTraceAll() *Client {
	c.trace = false
	return c
}

// EnableTraceAll is a global wrapper methods which delegated
// to the default client's EnableTraceAll.
func EnableTraceAll() *Client {
	return defaultClient.EnableTraceAll()
}

// EnableTraceAll enable trace for all requests.
func (c *Client) EnableTraceAll() *Client {
	c.trace = true
	return c
}

// SetCookieJar is a global wrapper methods which delegated
// to the default client's SetCookieJar.
func SetCookieJar(jar http.CookieJar) *Client {
	return defaultClient.SetCookieJar(jar)
}

// SetCookieJar set the `CookeJar` to the underlying `http.Client`.
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.httpClient.Jar = jar
	return c
}

// SetJsonMarshal is a global wrapper methods which delegated
// to the default client's SetJsonMarshal.
func SetJsonMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	return defaultClient.SetJsonMarshal(fn)
}

// SetJsonMarshal set the JSON marshal function which will be used
// to marshal request body.
func (c *Client) SetJsonMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	c.jsonMarshal = fn
	return c
}

// SetJsonUnmarshal is a global wrapper methods which delegated
// to the default client's SetJsonUnmarshal.
func SetJsonUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	return defaultClient.SetJsonUnmarshal(fn)
}

// SetJsonUnmarshal set the JSON unmarshal function which will be used
// to unmarshal response body.
func (c *Client) SetJsonUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	c.jsonUnmarshal = fn
	return c
}

// SetXmlMarshal is a global wrapper methods which delegated
// to the default client's SetXmlMarshal.
func SetXmlMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	return defaultClient.SetXmlMarshal(fn)
}

// SetXmlMarshal set the XML marshal function which will be used
// to marshal request body.
func (c *Client) SetXmlMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	c.xmlMarshal = fn
	return c
}

// SetXmlUnmarshal is a global wrapper methods which delegated
// to the default client's SetXmlUnmarshal.
func SetXmlUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	return defaultClient.SetXmlUnmarshal(fn)
}

// SetXmlUnmarshal set the XML unmarshal function which will be used
// to unmarshal response body.
func (c *Client) SetXmlUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	c.xmlUnmarshal = fn
	return c
}

// SetDialTLS is a global wrapper methods which delegated
// to the default client's SetDialTLS.
func SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	return defaultClient.SetDialTLS(fn)
}

// SetDialTLS set the customized `DialTLSContext` function to Transport.
// Make sure the returned `conn` implements TLSConn if you want your
// customized `conn` supports HTTP2.
func (c *Client) SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	c.t.DialTLSContext = fn
	return c
}

// SetDial is a global wrapper methods which delegated
// to the default client's SetDial.
func SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	return defaultClient.SetDial(fn)
}

// SetDial set the customized `DialContext` function to Transport.
func (c *Client) SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	c.t.DialContext = fn
	return c
}

// SetTLSHandshakeTimeout is a global wrapper methods which delegated
// to the default client's SetTLSHandshakeTimeout.
func SetTLSHandshakeTimeout(timeout time.Duration) *Client {
	return defaultClient.SetTLSHandshakeTimeout(timeout)
}

// SetTLSHandshakeTimeout set the TLS handshake timeout.
func (c *Client) SetTLSHandshakeTimeout(timeout time.Duration) *Client {
	c.t.TLSHandshakeTimeout = timeout
	return c
}

// EnableForceHTTP1 is a global wrapper methods which delegated
// to the default client's EnableForceHTTP1.
func EnableForceHTTP1() *Client {
	return defaultClient.EnableForceHTTP1()
}

// EnableForceHTTP1 enable force using HTTP1 (disabled by default).
func (c *Client) EnableForceHTTP1() *Client {
	c.t.ForceHttpVersion = HTTP1
	return c
}

// EnableForceHTTP2 is a global wrapper methods which delegated
// to the default client's EnableForceHTTP2.
func EnableForceHTTP2() *Client {
	return defaultClient.EnableForceHTTP2()
}

// EnableForceHTTP2 enable force using HTTP2 for https requests
// (disabled by default).
func (c *Client) EnableForceHTTP2() *Client {
	c.t.ForceHttpVersion = HTTP2
	return c
}

// DisableForceHttpVersion is a global wrapper methods which delegated
// to the default client's DisableForceHttpVersion.
func DisableForceHttpVersion() *Client {
	return defaultClient.DisableForceHttpVersion()
}

// DisableForceHttpVersion disable force using HTTP1 (disabled by default).
func (c *Client) DisableForceHttpVersion() *Client {
	c.t.ForceHttpVersion = ""
	return c
}

// DisableAllowGetMethodPayload is a global wrapper methods which delegated
// to the default client's DisableAllowGetMethodPayload.
func DisableAllowGetMethodPayload() *Client {
	return defaultClient.DisableAllowGetMethodPayload()
}

// DisableAllowGetMethodPayload disable sending GET method requests with body.
func (c *Client) DisableAllowGetMethodPayload() *Client {
	c.AllowGetMethodPayload = false
	return c
}

// EnableAllowGetMethodPayload is a global wrapper methods which delegated
// to the default client's EnableAllowGetMethodPayload.
func EnableAllowGetMethodPayload() *Client {
	return defaultClient.EnableAllowGetMethodPayload()
}

// EnableAllowGetMethodPayload allows sending GET method requests with body.
func (c *Client) EnableAllowGetMethodPayload() *Client {
	c.AllowGetMethodPayload = true
	return c
}

func (c *Client) isPayloadForbid(m string) bool {
	return (m == http.MethodGet && !c.AllowGetMethodPayload) || m == http.MethodHead || m == http.MethodOptions
}

// NewClient is the alias of C
func NewClient() *Client {
	return C()
}

// Clone copy and returns the Client
func (c *Client) Clone() *Client {
	t := c.t.Clone()
	t2, _ := http2ConfigureTransports(t)
	client := *c.httpClient
	client.Transport = t

	cc := *c
	cc.httpClient = &client
	cc.t = t
	cc.t2 = t2

	cc.Headers = cloneHeaders(c.Headers)
	cc.Cookies = cloneCookies(c.Cookies)
	cc.PathParams = cloneMap(c.PathParams)
	cc.QueryParams = cloneUrlValues(c.QueryParams)
	cc.FormData = cloneUrlValues(c.FormData)
	cc.beforeRequest = cloneRequestMiddleware(c.beforeRequest)
	cc.udBeforeRequest = cloneRequestMiddleware(c.udBeforeRequest)
	cc.afterResponse = cloneResponseMiddleware(c.afterResponse)
	cc.dumpOptions = c.dumpOptions.Clone()

	cc.log = c.log
	cc.jsonUnmarshal = c.jsonUnmarshal
	cc.jsonMarshal = c.jsonMarshal
	cc.xmlMarshal = c.xmlMarshal
	cc.xmlUnmarshal = c.xmlUnmarshal

	return &cc
}

// C create a new client.
func C() *Client {
	t := &Transport{
		ResponseOptions:       &ResponseOptions{},
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
		parseRequestHeader,
		parseRequestURL,
		parseRequestBody,
		parseRequestCookie,
	}
	afterResponse := []ResponseMiddleware{
		parseResponseBody,
		handleDownload,
	}
	c := &Client{
		beforeRequest: beforeRequest,
		afterResponse: afterResponse,
		log:           createDefaultLogger(),
		httpClient:    httpClient,
		t:             t,
		t2:            t2,
		jsonMarshal:   json.Marshal,
		jsonUnmarshal: json.Unmarshal,
		xmlMarshal:    xml.Marshal,
		xmlUnmarshal:  xml.Unmarshal,
	}
	httpClient.CheckRedirect = c.defaultCheckRedirect

	t.Debugf = func(format string, v ...interface{}) {
		if c.DebugLog {
			c.log.Debugf(format, v...)
		}
	}
	return c
}

func setupRequest(r *Request) {
	setRequestURL(r, r.URL)
	setRequestHeaderAndCookie(r)
	setTrace(r)
	setContext(r)
}

func (c *Client) do(r *Request) (resp *Response, err error) {
	resp = &Response{
		Request: r,
	}

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

	if c.DebugLog {
		c.log.Debugf("%s %s", r.RawRequest.Method, r.RawRequest.URL.String())
	}

	r.StartTime = time.Now()
	httpResponse, err := c.httpClient.Do(r.RawRequest)
	if err != nil {
		return
	}

	resp.Request = r
	resp.Response = httpResponse

	if !c.disableAutoReadResponse && !r.isSaveResponse { // auto read response body
		_, err = resp.ToBytes()
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

func setContext(r *Request) {
	if r.ctx != nil {
		r.RawRequest = r.RawRequest.WithContext(r.ctx)
	}
}

func setTrace(r *Request) {
	if r.trace == nil && r.client.trace {
		r.trace = &clientTrace{}
	}
	if r.trace != nil {
		r.ctx = r.trace.createContext(r.Context())
	}
}

func setRequestHeaderAndCookie(r *Request) {
	for k, vs := range r.Headers {
		for _, v := range vs {
			r.RawRequest.Header.Add(k, v)
		}
	}
	for _, cookie := range r.Cookies {
		r.RawRequest.AddCookie(cookie)
	}
}

func setRequestURL(r *Request, url string) error {
	// The host's colon:port should be normalized. See Issue 14836.
	u, err := urlpkg.Parse(url)
	if err != nil {
		return err
	}
	u.Host = removeEmptyPort(u.Host)
	if host := r.getHeader("Host"); host != "" {
		r.RawRequest.Host = host // Host header override
	} else {
		r.RawRequest.Host = u.Host
	}
	r.RawRequest.URL = u
	return nil
}
