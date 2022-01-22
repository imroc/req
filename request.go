package req

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"os"
	"strings"
)

// Request is the http request
type Request struct {
	URL            string
	PathParams     map[string]string
	QueryParams    urlpkg.Values
	error          error
	client         *Client
	httpRequest    *http.Request
	isSaveResponse bool
	output         io.WriteCloser
}

// New create a new request using the global default client.
func New() *Request {
	return defaultClient.R()
}

func (r *Request) SetOutputFile(file string) *Request {
	output, err := os.Create(file)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetOutput(output)
}

func (r *Request) SetOutput(output io.WriteCloser) *Request {
	r.output = output
	r.isSaveResponse = true
	return r
}

func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

func (r *Request) SetQueryParam(key, value string) *Request {
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	r.QueryParams.Set(key, value)
	return r
}

func (r *Request) SetPathParams(params map[string]string) *Request {
	for key, value := range params {
		r.SetPathParam(key, value)
	}
	return r
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

// Error return the underlying error, not nil if some error
// happend when constructing the request.
func (r *Request) Error() error {
	return r.error
}

func (r *Request) Send(method, url string) (*Response, error) {
	if r.error != nil {
		return nil, r.error
	}

	r.URL = url

	if method == "" {
		// We document that "" means "GET" for Request.Method, and people have
		// relied on that from NewRequest, so keep that working.
		// We still enforce validMethod for non-empty methods.
		method = "GET"
	}
	if !validMethod(method) {
		err := fmt.Errorf("net/http: invalid method %q", method)
		if err != nil {
			return nil, err
		}
	}
	r.httpRequest.Method = method

	for _, f := range r.client.beforeRequest {
		if err := f(r.client, r); err != nil {
			return nil, err
		}
	}

	// The host's colon:port should be normalized. See Issue 14836.
	u, err := urlpkg.Parse(r.URL)
	if err != nil {
		return nil, err
	}
	u.Host = removeEmptyPort(u.Host)
	r.httpRequest.URL = u
	r.httpRequest.Host = u.Host
	return r.execute()
}

// MustGet like Get, panic if error happens.
func (r *Request) MustGet(url string) *Response {
	resp, err := r.Get(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Get Send the request with GET method and specified url.
func (r *Request) Get(url string) (*Response, error) {
	return r.Send(http.MethodGet, url)
}

// MustPost like Post, panic if error happens.
func (r *Request) MustPost(url string) *Response {
	resp, err := r.Post(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Post Send the request with POST method and specified url.
func (r *Request) Post(url string) (*Response, error) {
	return r.Send(http.MethodPost, url)
}

// MustPut like Put, panic if error happens.
func (r *Request) MustPut(url string) *Response {
	resp, err := r.Put(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Put Send the request with Put method and specified url.
func (r *Request) Put(url string) (*Response, error) {
	return r.Send(http.MethodPut, url)
}

// MustPatch like Patch, panic if error happens.
func (r *Request) MustPatch(url string) *Response {
	resp, err := r.Patch(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Patch Send the request with PATCH method and specified url.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Send(http.MethodPatch, url)
}

// MustDelete like Delete, panic if error happens.
func (r *Request) MustDelete(url string) *Response {
	resp, err := r.Delete(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Delete Send the request with DELETE method and specified url.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Send(http.MethodDelete, url)
}

// MustOptions like Options, panic if error happens.
func (r *Request) MustOptions(url string) *Response {
	resp, err := r.Options(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Options Send the request with OPTIONS method and specified url.
func (r *Request) Options(url string) (*Response, error) {
	return r.Send(http.MethodOptions, url)
}

// MustHead like Head, panic if error happens.
func (r *Request) MustHead(url string) *Response {
	resp, err := r.Send(http.MethodHead, url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Head Send the request with HEAD method and specified url.
func (r *Request) Head(url string) (*Response, error) {
	return r.Send(http.MethodHead, url)
}

// SetBody set the request body.
func (r *Request) SetBody(body interface{}) *Request {
	if body == nil {
		return r
	}
	switch b := body.(type) {
	case io.ReadCloser:
		r.httpRequest.Body = b
	case io.Reader:
		r.httpRequest.Body = ioutil.NopCloser(b)
	case []byte:
		r.SetBodyBytes(b)
	case string:
		r.SetBodyString(b)
	}
	return r
}

// SetBodyBytes set the request body as []byte.
func (r *Request) SetBodyBytes(body []byte) *Request {
	r.httpRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r
}

// SetBodyString set the request body as string.
func (r *Request) SetBodyString(body string) *Request {
	r.httpRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r
}

// SetBodyJsonString set the request body as string and set Content-Type header
// as "application/json; charset=UTF-8"
func (r *Request) SetBodyJsonString(body string) *Request {
	r.httpRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	r.SetContentType(CONTENT_TYPE_APPLICATION_JSON_UTF8)
	return r
}

// SetBodyJsonBytes set the request body as []byte and set Content-Type header
// as "application/json; charset=UTF-8"
func (r *Request) SetBodyJsonBytes(body []byte) *Request {
	r.httpRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	r.SetContentType(CONTENT_TYPE_APPLICATION_JSON_UTF8)
	return r
}

// SetBodyJsonMarshal set the request body that marshaled from object, and
// set Content-Type header as "application/json; charset=UTF-8"
func (r *Request) SetBodyJsonMarshal(v interface{}) *Request {
	b, err := json.Marshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetBodyBytes(b)
}

func (r *Request) SetContentType(contentType string) *Request {
	r.httpRequest.Header.Set("Content-Type", contentType)
	return r
}

func (r *Request) execute() (resp *Response, err error) {
	if r.error != nil {
		return nil, r.error
	}
	for k, v := range r.client.commonHeader {
		if r.httpRequest.Header.Get(k) == "" {
			r.httpRequest.Header.Set(k, v)
		}
	}
	logf(r.client.log, "%s %s", r.httpRequest.Method, r.httpRequest.URL.String())
	httpResponse, err := r.client.httpClient.Do(r.httpRequest)
	if err != nil {
		return
	}
	resp = &Response{
		request:  r,
		Response: httpResponse,
	}
	if r.client.t.ResponseOptions != nil && r.client.t.ResponseOptions.AutoDiscard {
		err = resp.Discard()
	}
	return
}
