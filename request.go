package req

import (
	"bytes"
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/imroc/req/v2/internal/util"
	"io"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"os"
	"strings"
	"time"
)

// Request is the http request
type Request struct {
	URL         string
	PathParams  map[string]string
	QueryParams urlpkg.Values
	FormData    urlpkg.Values
	Headers     http.Header
	Cookies     []*http.Cookie
	Result      interface{}
	Error       interface{}
	error       error
	client      *Client
	RawRequest  *http.Request
	StartTime   time.Time

	ctx            context.Context
	isMultiPart    bool
	uploadFiles    []*uploadFile
	uploadReader   []io.ReadCloser
	outputFile     string
	isSaveResponse bool
	output         io.WriteCloser
	trace          *clientTrace
}

func (r *Request) TraceInfo() TraceInfo {
	ct := r.trace

	if ct == nil {
		return TraceInfo{}
	}

	ti := TraceInfo{
		DNSLookupTime:     ct.dnsDone.Sub(ct.dnsStart),
		TLSHandshakeTime:  ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart),
		FirstResponseTime: ct.gotFirstResponseByte.Sub(ct.gotConn),
		IsConnReused:      ct.gotConnInfo.Reused,
		IsConnWasIdle:     ct.gotConnInfo.WasIdle,
		ConnIdleTime:      ct.gotConnInfo.IdleTime,
	}

	// Calculate the total time accordingly,
	// when connection is reused
	if ct.gotConnInfo.Reused {
		ti.TotalTime = ct.endTime.Sub(ct.getConn)
	} else {
		ti.TotalTime = ct.endTime.Sub(ct.dnsStart)
	}

	// Only calculate on successful connections
	if !ct.connectDone.IsZero() {
		ti.TCPConnectTime = ct.connectDone.Sub(ct.dnsDone)
	}

	// Only calculate on successful connections
	if !ct.gotConn.IsZero() {
		ti.ConnectTime = ct.gotConn.Sub(ct.getConn)
	}

	// Only calculate on successful connections
	if !ct.gotFirstResponseByte.IsZero() {
		ti.ResponseTime = ct.endTime.Sub(ct.gotFirstResponseByte)
	}

	// Capture remote address info when connection is non-nil
	if ct.gotConnInfo.Conn != nil {
		ti.RemoteAddr = ct.gotConnInfo.Conn.RemoteAddr()
	}

	return ti
}

func SetFormDataFromValues(data urlpkg.Values) *Request {
	return defaultClient.R().SetFormDataFromValues(data)
}

func (r *Request) SetFormDataFromValues(data urlpkg.Values) *Request {
	if r.FormData == nil {
		r.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		for _, kv := range v {
			r.FormData.Add(k, kv)
		}
	}
	return r
}

func SetFormData(data map[string]string) *Request {
	return defaultClient.R().SetFormData(data)
}

func (r *Request) SetFormData(data map[string]string) *Request {
	if r.FormData == nil {
		r.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		r.FormData.Set(k, v)
	}
	return r
}

func SetCookie(hc *http.Cookie) *Request {
	return defaultClient.R().SetCookie(hc)
}

func (r *Request) SetCookie(hc *http.Cookie) *Request {
	r.Cookies = append(r.Cookies, hc)
	return r
}

func SetCookies(rs []*http.Cookie) *Request {
	return defaultClient.R().SetCookies(rs)
}

func (r *Request) SetCookies(rs []*http.Cookie) *Request {
	r.Cookies = append(r.Cookies, rs...)
	return r
}

func SetQueryString(query string) *Request {
	return defaultClient.R().SetQueryString(query)
}

func (r *Request) SetQueryString(query string) *Request {
	params, err := urlpkg.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		for p, v := range params {
			for _, pv := range v {
				r.QueryParams.Add(p, pv)
			}
		}
	} else {
		r.client.log.Errorf("%v", err)
	}
	return r
}

func SetFileReader(paramName, filePath string, reader io.Reader) *Request {
	return defaultClient.R().SetFileReader(paramName, filePath, reader)
}

func (r *Request) SetFileReader(paramName, filePath string, reader io.Reader) *Request {
	r.isMultiPart = true
	r.uploadFiles = append(r.uploadFiles, &uploadFile{
		ParamName: paramName,
		FilePath:  filePath,
		Reader:    reader,
	})
	return r
}

func SetFiles(files map[string]string) *Request {
	return defaultClient.R().SetFiles(files)
}

func (r *Request) SetFiles(files map[string]string) *Request {
	for k, v := range files {
		r.SetFile(k, v)
	}
	return r
}

func SetFile(paramName, filePath string) *Request {
	return defaultClient.R().SetFile(paramName, filePath)
}

func (r *Request) SetFile(paramName, filePath string) *Request {
	r.isMultiPart = true
	file, err := os.Open(filePath)
	if err != nil {
		r.client.log.Errorf("failed to open %s: %v", filePath, err)
		r.appendError(err)
		return r
	}
	r.uploadFiles = append(r.uploadFiles, &uploadFile{
		ParamName: paramName,
		FilePath:  filePath,
		Reader:    file,
	})
	return r
}

func SetResult(result interface{}) *Request {
	return defaultClient.R().SetResult(result)
}

func (r *Request) SetResult(result interface{}) *Request {
	r.Result = util.GetPointer(result)
	return r
}

func SetError(error interface{}) *Request {
	return defaultClient.R().SetError(error)
}

func (r *Request) SetError(error interface{}) *Request {
	r.Error = util.GetPointer(error)
	return r
}

func SetBearerAuthToken(token string) *Request {
	return defaultClient.R().SetBearerAuthToken(token)
}

func (r *Request) SetBearerAuthToken(token string) *Request {
	return r.SetHeader("Authorization", "Bearer "+token)
}

func SetBasicAuth(username, password string) *Request {
	return defaultClient.R().SetBasicAuth(username, password)
}

func (r *Request) SetBasicAuth(username, password string) *Request {
	return r.SetHeader("Authorization", util.BasicAuthHeaderValue(username, password))
}

func SetHeaders(hdrs map[string]string) *Request {
	return defaultClient.R().SetHeaders(hdrs)
}

func (r *Request) SetHeaders(hdrs map[string]string) *Request {
	for k, v := range hdrs {
		r.SetHeader(k, v)
	}
	return r
}

func SetHeader(key, value string) *Request {
	return defaultClient.R().SetHeader(key, value)
}

// SetHeader set the common header for all requests.
func (r *Request) SetHeader(key, value string) *Request {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set(key, value)
	return r
}

func SetOutputFile(file string) *Request {
	return defaultClient.R().SetOutputFile(file)
}

func (r *Request) SetOutputFile(file string) *Request {
	r.isSaveResponse = true
	r.outputFile = file
	return r
}

func SetOutput(output io.WriteCloser) *Request {
	return defaultClient.R().SetOutput(output)
}

func (r *Request) SetOutput(output io.WriteCloser) *Request {
	r.output = output
	r.isSaveResponse = true
	return r
}

func SetQueryParams(params map[string]string) *Request {
	return defaultClient.R().SetQueryParams(params)
}

func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

func SetQueryParam(key, value string) *Request {
	return defaultClient.R().SetQueryParam(key, value)
}

func (r *Request) SetQueryParam(key, value string) *Request {
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	r.QueryParams.Set(key, value)
	return r
}

func SetPathParams(params map[string]string) *Request {
	return defaultClient.R().SetPathParams(params)
}

func (r *Request) SetPathParams(params map[string]string) *Request {
	for key, value := range params {
		r.SetPathParam(key, value)
	}
	return r
}

func SetPathParam(key, value string) *Request {
	return defaultClient.R().SetPathParam(key, value)
}

func (r *Request) SetPathParam(key, value string) *Request {
	if r.PathParams == nil {
		r.PathParams = make(map[string]string)
	}
	r.PathParams[key] = value
	return r
}

func (r *Request) appendError(err error) {
	r.error = multierror.Append(r.error, err)
}

func (r *Request) Send(method, url string) (*Response, error) {
	if r.error != nil {
		return &Response{}, r.error
	}
	r.RawRequest.Method = method
	r.URL = url
	return r.client.do(r)
}

func MustGet(url string) *Response {
	return defaultClient.R().MustGet(url)
}

// MustGet like Get, panic if error happens.
func (r *Request) MustGet(url string) *Response {
	resp, err := r.Get(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func Get(url string) (*Response, error) {
	return defaultClient.R().Get(url)
}

// Get Send the request with GET method and specified url.
func (r *Request) Get(url string) (*Response, error) {
	return r.Send(http.MethodGet, url)
}

func MustPost(url string) *Response {
	return defaultClient.R().MustPost(url)
}

// MustPost like Post, panic if error happens.
func (r *Request) MustPost(url string) *Response {
	resp, err := r.Post(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func Post(url string) (*Response, error) {
	return defaultClient.R().Post(url)
}

// Post Send the request with POST method and specified url.
func (r *Request) Post(url string) (*Response, error) {
	return r.Send(http.MethodPost, url)
}

func MustPut(url string) *Response {
	return defaultClient.R().MustPut(url)
}

// MustPut like Put, panic if error happens.
func (r *Request) MustPut(url string) *Response {
	resp, err := r.Put(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func Put(url string) (*Response, error) {
	return defaultClient.R().Put(url)
}

// Put Send the request with Put method and specified url.
func (r *Request) Put(url string) (*Response, error) {
	return r.Send(http.MethodPut, url)
}

func MustPatch(url string) *Response {
	return defaultClient.R().MustPatch(url)
}

// MustPatch like Patch, panic if error happens.
func (r *Request) MustPatch(url string) *Response {
	resp, err := r.Patch(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func Patch(url string) (*Response, error) {
	return defaultClient.R().Patch(url)
}

// Patch Send the request with PATCH method and specified url.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Send(http.MethodPatch, url)
}

func MustDelete(url string) *Response {
	return defaultClient.R().MustDelete(url)
}

// MustDelete like Delete, panic if error happens.
func (r *Request) MustDelete(url string) *Response {
	resp, err := r.Delete(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func Delete(url string) (*Response, error) {
	return defaultClient.R().Delete(url)
}

// Delete Send the request with DELETE method and specified url.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Send(http.MethodDelete, url)
}

func MustOptions(url string) *Response {
	return defaultClient.R().MustOptions(url)
}

// MustOptions like Options, panic if error happens.
func (r *Request) MustOptions(url string) *Response {
	resp, err := r.Options(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func Options(url string) (*Response, error) {
	return defaultClient.R().Options(url)
}

// Options Send the request with OPTIONS method and specified url.
func (r *Request) Options(url string) (*Response, error) {
	return r.Send(http.MethodOptions, url)
}

func MustHead(url string) *Response {
	return defaultClient.R().MustHead(url)
}

// MustHead like Head, panic if error happens.
func (r *Request) MustHead(url string) *Response {
	resp, err := r.Send(http.MethodHead, url)
	if err != nil {
		panic(err)
	}
	return resp
}

func Head(url string) (*Response, error) {
	return defaultClient.R().Head(url)
}

// Head Send the request with HEAD method and specified url.
func (r *Request) Head(url string) (*Response, error) {
	return r.Send(http.MethodHead, url)
}

func SetBody(body interface{}) *Request {
	return defaultClient.R().SetBody(body)
}

// SetBody set the request body.
func (r *Request) SetBody(body interface{}) *Request {
	if body == nil {
		return r
	}
	switch b := body.(type) {
	case io.ReadCloser:
		r.RawRequest.Body = b
	case io.Reader:
		r.RawRequest.Body = ioutil.NopCloser(b)
	case []byte:
		r.SetBodyBytes(b)
	case string:
		r.SetBodyString(b)
	}
	return r
}

func SetBodyBytes(body []byte) *Request {
	return defaultClient.R().SetBodyBytes(body)
}

// SetBodyBytes set the request body as []byte.
func (r *Request) SetBodyBytes(body []byte) *Request {
	r.RawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r
}

func SetBodyString(body string) *Request {
	return defaultClient.R().SetBodyString(body)
}

// SetBodyString set the request body as string.
func (r *Request) SetBodyString(body string) *Request {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r
}

func SetBodyJsonString(body string) *Request {
	return defaultClient.R().SetBodyJsonString(body)
}

// SetBodyJsonString set the request body as string and set Content-Type header
// as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonString(body string) *Request {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	r.SetContentType(jsonContentType)
	return r
}

func SetBodyJsonBytes(body []byte) *Request {
	return defaultClient.R().SetBodyJsonBytes(body)
}

// SetBodyJsonBytes set the request body as []byte and set Content-Type header
// as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonBytes(body []byte) *Request {
	r.RawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	r.SetContentType(jsonContentType)
	return r
}

func SetBodyJsonMarshal(v interface{}) *Request {
	return defaultClient.R().SetBodyJsonMarshal(v)
}

// SetBodyJsonMarshal set the request body that marshaled from object, and
// set Content-Type header as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonMarshal(v interface{}) *Request {
	b, err := r.client.JSONMarshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetBodyBytes(b)
}

func SetBodyXmlMarshal(v interface{}) *Request {
	return defaultClient.R().SetBodyXmlMarshal(v)
}

func (r *Request) SetBodyXmlMarshal(v interface{}) *Request {
	b, err := r.client.XMLMarshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetBodyBytes(b)
}

func SetContentType(contentType string) *Request {
	return defaultClient.R().SetContentType(contentType)
}

func (r *Request) SetContentType(contentType string) *Request {
	r.RawRequest.Header.Set(hdrContentTypeKey, contentType)
	return r
}

// Context method returns the Context if its already set in request
// otherwise it creates new one using `context.Background()`.
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	return r.ctx
}

func SetContext(ctx context.Context) *Request {
	return defaultClient.R().SetContext(ctx)
}

// SetContext method sets the context.Context for current Request. It allows
// to interrupt the request execution if ctx.Done() channel is closed.
// See https://blog.golang.org/context article and the "context" package
// documentation.
func (r *Request) SetContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func EnableTrace(enable bool) *Request {
	return defaultClient.R().EnableTrace(enable)
}

func (r *Request) EnableTrace(enable bool) *Request {
	if enable {
		if r.trace == nil {
			r.trace = &clientTrace{}
		}
	} else {
		r.trace = nil
	}
	return r
}
