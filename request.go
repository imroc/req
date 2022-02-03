package req

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/imroc/req/v2/internal/util"
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
	Body        []byte

	marshalBody    interface{}
	ctx            context.Context
	isMultiPart    bool
	uploadFiles    []*uploadFile
	uploadReader   []io.ReadCloser
	outputFile     string
	isSaveResponse bool
	output         io.Writer
	trace          *clientTrace
}

// TraceInfo returns the trace information, only available when trace is enabled.
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

// SetFormDataFromValues is a global wrapper methods which delegated
// to the default client, create a request and SetFormDataFromValues for request.
func SetFormDataFromValues(data urlpkg.Values) *Request {
	return defaultClient.R().SetFormDataFromValues(data)
}

// SetFormDataFromValues set the form data from url.Values, not used if method not allow payload.
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

// SetFormData is a global wrapper methods which delegated
// to the default client, create a request and SetFormData for request.
func SetFormData(data map[string]string) *Request {
	return defaultClient.R().SetFormData(data)
}

// SetFormData set the form data from map, not used if method not allow payload.
func (r *Request) SetFormData(data map[string]string) *Request {
	if r.FormData == nil {
		r.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		r.FormData.Set(k, v)
	}
	return r
}

// SetCookies is a global wrapper methods which delegated
// to the default client, create a request and SetCookies for request.
func SetCookies(cookies ...*http.Cookie) *Request {
	return defaultClient.R().SetCookies(cookies...)
}

// SetCookies set cookies at request level.
func (r *Request) SetCookies(cookies ...*http.Cookie) *Request {
	r.Cookies = append(r.Cookies, cookies...)
	return r
}

// SetQueryString is a global wrapper methods which delegated
// to the default client, create a request and SetQueryString for request.
func SetQueryString(query string) *Request {
	return defaultClient.R().SetQueryString(query)
}

// SetQueryString set URL query parameters using the raw query string.
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

// SetFileReader is a global wrapper methods which delegated
// to the default client, create a request and SetFileReader for request.
func SetFileReader(paramName, filePath string, reader io.Reader) *Request {
	return defaultClient.R().SetFileReader(paramName, filePath, reader)
}

// SetFileReader sets up a multipart form with a reader to upload file.
func (r *Request) SetFileReader(paramName, filePath string, reader io.Reader) *Request {
	r.isMultiPart = true
	r.uploadFiles = append(r.uploadFiles, &uploadFile{
		ParamName: paramName,
		FilePath:  filePath,
		Reader:    reader,
	})
	return r
}

// SetFiles is a global wrapper methods which delegated
// to the default client, create a request and SetFiles for request.
func SetFiles(files map[string]string) *Request {
	return defaultClient.R().SetFiles(files)
}

// SetFiles sets up a multipart form from a map, which key is the param
// name, value is the file path.
func (r *Request) SetFiles(files map[string]string) *Request {
	for k, v := range files {
		r.SetFile(k, v)
	}
	return r
}

// SetFile is a global wrapper methods which delegated
// to the default client, create a request and SetFile for request.
func SetFile(paramName, filePath string) *Request {
	return defaultClient.R().SetFile(paramName, filePath)
}

// SetFile sets up a multipart form, read file from filePath automatically to upload.
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

// SetResult is a global wrapper methods which delegated
// to the default client, create a request and SetResult for request.
func SetResult(result interface{}) *Request {
	return defaultClient.R().SetResult(result)
}

// SetResult set the result that response body will be unmarshaled to if
// request is success (status `code >= 200 and <= 299`).
func (r *Request) SetResult(result interface{}) *Request {
	r.Result = util.GetPointer(result)
	return r
}

// SetError is a global wrapper methods which delegated
// to the default client, create a request and SetError for request.
func SetError(error interface{}) *Request {
	return defaultClient.R().SetError(error)
}

// SetError set the result that response body will be unmarshaled to if
// request is error ( status `code >= 400`).
func (r *Request) SetError(error interface{}) *Request {
	r.Error = util.GetPointer(error)
	return r
}

// SetBearerAuthToken is a global wrapper methods which delegated
// to the default client, create a request and SetBearerAuthToken for request.
func SetBearerAuthToken(token string) *Request {
	return defaultClient.R().SetBearerAuthToken(token)
}

// SetBearerAuthToken set the bearer auth token at request level.
func (r *Request) SetBearerAuthToken(token string) *Request {
	return r.SetHeader("Authorization", "Bearer "+token)
}

// SetBasicAuth is a global wrapper methods which delegated
// to the default client, create a request and SetBasicAuth for request.
func SetBasicAuth(username, password string) *Request {
	return defaultClient.R().SetBasicAuth(username, password)
}

// SetBasicAuth set the basic auth at request level.
func (r *Request) SetBasicAuth(username, password string) *Request {
	return r.SetHeader("Authorization", util.BasicAuthHeaderValue(username, password))
}

// SetHeaders is a global wrapper methods which delegated
// to the default client, create a request and SetHeaders for request.
func SetHeaders(hdrs map[string]string) *Request {
	return defaultClient.R().SetHeaders(hdrs)
}

// SetHeaders set the header at request level.
func (r *Request) SetHeaders(hdrs map[string]string) *Request {
	for k, v := range hdrs {
		r.SetHeader(k, v)
	}
	return r
}

// SetHeader is a global wrapper methods which delegated
// to the default client, create a request and SetHeader for request.
func SetHeader(key, value string) *Request {
	return defaultClient.R().SetHeader(key, value)
}

// SetHeader set a header at request level.
func (r *Request) SetHeader(key, value string) *Request {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set(key, value)
	return r
}

// SetOutputFile is a global wrapper methods which delegated
// to the default client, create a request and SetOutputFile for request.
func SetOutputFile(file string) *Request {
	return defaultClient.R().SetOutputFile(file)
}

// SetOutputFile the file that response body will be downloaded to.
func (r *Request) SetOutputFile(file string) *Request {
	r.isSaveResponse = true
	r.outputFile = file
	return r
}

// SetOutput is a global wrapper methods which delegated
// to the default client, create a request and SetOutput for request.
func SetOutput(output io.Writer) *Request {
	return defaultClient.R().SetOutput(output)
}

// SetOutput the io.Writer that response body will be downloaded to.
func (r *Request) SetOutput(output io.Writer) *Request {
	r.output = output
	r.isSaveResponse = true
	return r
}

// SetQueryParams is a global wrapper methods which delegated
// to the default client, create a request and SetQueryParams for request.
func SetQueryParams(params map[string]string) *Request {
	return defaultClient.R().SetQueryParams(params)
}

// SetQueryParams sets the URL query parameters with a map at client level.
func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

// SetQueryParam is a global wrapper methods which delegated
// to the default client, create a request and SetQueryParam for request.
func SetQueryParam(key, value string) *Request {
	return defaultClient.R().SetQueryParam(key, value)
}

// SetQueryParam set an URL query parameter with a key-value
// pair at request level.
func (r *Request) SetQueryParam(key, value string) *Request {
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	r.QueryParams.Set(key, value)
	return r
}

// AddQueryParam is a global wrapper methods which delegated
// to the default client, create a request and AddQueryParam for request.
func AddQueryParam(key, value string) *Request {
	return defaultClient.R().AddQueryParam(key, value)
}

// AddQueryParam add a URL query parameter with a key-value
// pair at request level.
func (r *Request) AddQueryParam(key, value string) *Request {
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	r.QueryParams.Add(key, value)
	return r
}

// SetPathParams is a global wrapper methods which delegated
// to the default client, create a request and SetPathParams for request.
func SetPathParams(params map[string]string) *Request {
	return defaultClient.R().SetPathParams(params)
}

// SetPathParams sets the URL path parameters from a map at request level.
func (r *Request) SetPathParams(params map[string]string) *Request {
	for key, value := range params {
		r.SetPathParam(key, value)
	}
	return r
}

// SetPathParam is a global wrapper methods which delegated
// to the default client, create a request and SetPathParam for request.
func SetPathParam(key, value string) *Request {
	return defaultClient.R().SetPathParam(key, value)
}

// SetPathParam sets the URL path parameters from a key-value paire at request level.
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

// Send sends the http request.
func (r *Request) Send(method, url string) (*Response, error) {
	if r.error != nil {
		return &Response{}, r.error
	}
	r.RawRequest.Method = method
	r.URL = url
	return r.client.do(r)
}

// MustGet is a global wrapper methods which delegated
// to the default client, create a request and MustGet for request.
func MustGet(url string) *Response {
	return defaultClient.R().MustGet(url)
}

// MustGet like Get, panic if error happens, should only be used to test without error handling.
func (r *Request) MustGet(url string) *Response {
	resp, err := r.Get(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Get is a global wrapper methods which delegated
// to the default client, create a request and Get for request.
func Get(url string) (*Response, error) {
	return defaultClient.R().Get(url)
}

// Get Send the request with GET method and specified url.
func (r *Request) Get(url string) (*Response, error) {
	return r.Send(http.MethodGet, url)
}

// MustPost is a global wrapper methods which delegated
// to the default client, create a request and Get for request.
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

// Post is a global wrapper methods which delegated
// to the default client, create a request and Post for request.
func Post(url string) (*Response, error) {
	return defaultClient.R().Post(url)
}

// Post Send the request with POST method and specified url.
func (r *Request) Post(url string) (*Response, error) {
	return r.Send(http.MethodPost, url)
}

// MustPut is a global wrapper methods which delegated
// to the default client, create a request and MustPut for request.
func MustPut(url string) *Response {
	return defaultClient.R().MustPut(url)
}

// MustPut like Put, panic if error happens, should only be used to test without error handling.
func (r *Request) MustPut(url string) *Response {
	resp, err := r.Put(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Put is a global wrapper methods which delegated
// to the default client, create a request and Put for request.
func Put(url string) (*Response, error) {
	return defaultClient.R().Put(url)
}

// Put Send the request with Put method and specified url.
func (r *Request) Put(url string) (*Response, error) {
	return r.Send(http.MethodPut, url)
}

// MustPatch is a global wrapper methods which delegated
// to the default client, create a request and MustPatch for request.
func MustPatch(url string) *Response {
	return defaultClient.R().MustPatch(url)
}

// MustPatch like Patch, panic if error happens, should only be used to test without error handling.
func (r *Request) MustPatch(url string) *Response {
	resp, err := r.Patch(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Patch is a global wrapper methods which delegated
// to the default client, create a request and Patch for request.
func Patch(url string) (*Response, error) {
	return defaultClient.R().Patch(url)
}

// Patch Send the request with PATCH method and specified url.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Send(http.MethodPatch, url)
}

// MustDelete is a global wrapper methods which delegated
// to the default client, create a request and MustDelete for request.
func MustDelete(url string) *Response {
	return defaultClient.R().MustDelete(url)
}

// MustDelete like Delete, panic if error happens, should only be used to test without error handling.
func (r *Request) MustDelete(url string) *Response {
	resp, err := r.Delete(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Delete is a global wrapper methods which delegated
// to the default client, create a request and Delete for request.
func Delete(url string) (*Response, error) {
	return defaultClient.R().Delete(url)
}

// Delete Send the request with DELETE method and specified url.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Send(http.MethodDelete, url)
}

// MustOptions is a global wrapper methods which delegated
// to the default client, create a request and MustOptions for request.
func MustOptions(url string) *Response {
	return defaultClient.R().MustOptions(url)
}

// MustOptions like Options, panic if error happens, should only be used to test without error handling.
func (r *Request) MustOptions(url string) *Response {
	resp, err := r.Options(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Options is a global wrapper methods which delegated
// to the default client, create a request and Options for request.
func Options(url string) (*Response, error) {
	return defaultClient.R().Options(url)
}

// Options Send the request with OPTIONS method and specified url.
func (r *Request) Options(url string) (*Response, error) {
	return r.Send(http.MethodOptions, url)
}

// MustHead is a global wrapper methods which delegated
// to the default client, create a request and MustHead for request.
func MustHead(url string) *Response {
	return defaultClient.R().MustHead(url)
}

// MustHead like Head, panic if error happens, should only be used to test without error handling.
func (r *Request) MustHead(url string) *Response {
	resp, err := r.Send(http.MethodHead, url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Head is a global wrapper methods which delegated
// to the default client, create a request and Head for request.
func Head(url string) (*Response, error) {
	return defaultClient.R().Head(url)
}

// Head Send the request with HEAD method and specified url.
func (r *Request) Head(url string) (*Response, error) {
	return r.Send(http.MethodHead, url)
}

// SetBody is a global wrapper methods which delegated
// to the default client, create a request and SetBody for request.
func SetBody(body interface{}) *Request {
	return defaultClient.R().SetBody(body)
}

// SetBody set the request body, accepts string, []byte, io.Reader, map and struct.
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
	default:
		r.marshalBody = body
	}
	return r
}

// SetBodyBytes is a global wrapper methods which delegated
// to the default client, create a request and SetBodyBytes for request.
func SetBodyBytes(body []byte) *Request {
	return defaultClient.R().SetBodyBytes(body)
}

// SetBodyBytes set the request body as []byte.
func (r *Request) SetBodyBytes(body []byte) *Request {
	r.RawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r
}

// SetBodyString is a global wrapper methods which delegated
// to the default client, create a request and SetBodyString for request.
func SetBodyString(body string) *Request {
	return defaultClient.R().SetBodyString(body)
}

// SetBodyString set the request body as string.
func (r *Request) SetBodyString(body string) *Request {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r
}

// SetBodyJsonString is a global wrapper methods which delegated
// to the default client, create a request and SetBodyJsonString for request.
func SetBodyJsonString(body string) *Request {
	return defaultClient.R().SetBodyJsonString(body)
}

// SetBodyJsonString set the request body as string and set Content-Type header
// as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonString(body string) *Request {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r.SetContentType(jsonContentType)
}

// SetBodyJsonBytes is a global wrapper methods which delegated
// to the default client, create a request and SetBodyJsonBytes for request.
func SetBodyJsonBytes(body []byte) *Request {
	return defaultClient.R().SetBodyJsonBytes(body)
}

// SetBodyJsonBytes set the request body as []byte and set Content-Type header
// as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonBytes(body []byte) *Request {
	r.RawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r.SetContentType(jsonContentType)
}

// SetBodyJsonMarshal is a global wrapper methods which delegated
// to the default client, create a request and SetBodyJsonMarshal for request.
func SetBodyJsonMarshal(v interface{}) *Request {
	return defaultClient.R().SetBodyJsonMarshal(v)
}

// SetBodyJsonMarshal set the request body that marshaled from object, and
// set Content-Type header as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonMarshal(v interface{}) *Request {
	b, err := r.client.jsonMarshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetContentType(jsonContentType).SetBodyBytes(b)
}

// SetBodyXmlString is a global wrapper methods which delegated
// to the default client, create a request and SetBodyXmlString for request.
func SetBodyXmlString(body string) *Request {
	return defaultClient.R().SetBodyXmlString(body)
}

// SetBodyXmlString set the request body as string and set Content-Type header
// as "text/xml; charset=utf-8"
func (r *Request) SetBodyXmlString(body string) *Request {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r.SetContentType(xmlContentType)
}

// SetBodyXmlBytes is a global wrapper methods which delegated
// to the default client, create a request and SetBodyXmlBytes for request.
func SetBodyXmlBytes(body []byte) *Request {
	return defaultClient.R().SetBodyXmlBytes(body)
}

// SetBodyXmlBytes set the request body as []byte and set Content-Type header
// as "text/xml; charset=utf-8"
func (r *Request) SetBodyXmlBytes(body []byte) *Request {
	r.RawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r.SetContentType(xmlContentType)
}

// SetBodyXmlMarshal is a global wrapper methods which delegated
// to the default client, create a request and SetBodyXmlMarshal for request.
func SetBodyXmlMarshal(v interface{}) *Request {
	return defaultClient.R().SetBodyXmlMarshal(v)
}

// SetBodyXmlMarshal set the request body that marshaled from object, and
// set Content-Type header as "text/xml; charset=utf-8"
func (r *Request) SetBodyXmlMarshal(v interface{}) *Request {
	b, err := r.client.xmlMarshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetContentType(xmlContentType).SetBodyBytes(b)
}

// SetContentType is a global wrapper methods which delegated
// to the default client, create a request and SetContentType for request.
func SetContentType(contentType string) *Request {
	return defaultClient.R().SetContentType(contentType)
}

// SetContentType set the `Content-Type` for the request.
func (r *Request) SetContentType(contentType string) *Request {
	r.SetHeader(hdrContentTypeKey, contentType)
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

// SetContext is a global wrapper methods which delegated
// to the default client, create a request and SetContext for request.
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

// EnableTrace is a global wrapper methods which delegated
// to the default client, create a request and EnableTrace for request.
func EnableTrace(enable bool) *Request {
	return defaultClient.R().EnableTrace(enable)
}

// EnableTrace enables trace if set to true.
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

// String returns the request body as a string
// if the client has SaveRequestBody enabled
func (r *Request) String() string {
	return string(r.Body)
}
