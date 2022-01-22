package req

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"strings"
)

// Request is the http request
type Request struct {
	error       error
	client      *Client
	httpRequest *http.Request
}

// New create a new request using the global default client.
func New() *Request {
	return defaultClient.R()
}

func (r *Request) appendError(err error) {
	r.error = multierror.Append(r.error, err)
}

// Error return the underlying error, not nil if some error
// happend when constructing the request.
func (r *Request) Error() error {
	return r.error
}

// Method set the http request method.
func (r *Request) Method(method string) *Request {
	if method == "" {
		// We document that "" means "GET" for Request.Method, and people have
		// relied on that from NewRequest, so keep that working.
		// We still enforce validMethod for non-empty methods.
		method = "GET"
	}
	if !validMethod(method) {
		err := fmt.Errorf("net/http: invalid method %q", method)
		if err != nil {
			r.appendError(err)
		}
	}
	r.httpRequest.Method = method
	r.httpRequest = r.httpRequest.WithContext(context.Background())
	return r
}

// URL set the http request url.
func (r *Request) URL(url string) *Request {
	u, err := urlpkg.Parse(url)
	if err != nil {
		r.appendError(err)
		return r
	}
	// The host's colon:port should be normalized. See Issue 14836.
	u.Host = removeEmptyPort(u.Host)
	r.httpRequest.URL = u
	r.httpRequest.Host = u.Host
	return r
}

func (r *Request) send(method, url string) (*Response, error) {
	return r.Method(method).URL(url).Send()
}

// MustGet like Get, panic if error happens.
func (r *Request) MustGet(url string) *Response {
	resp, err := r.Get(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Get send the request with GET method and specified url.
func (r *Request) Get(url string) (*Response, error) {
	return r.send(http.MethodGet, url)
}

// MustPost like Post, panic if error happens.
func (r *Request) MustPost(url string) *Response {
	resp, err := r.Post(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Post send the request with POST method and specified url.
func (r *Request) Post(url string) (*Response, error) {
	return r.send(http.MethodPost, url)
}

// MustPut like Put, panic if error happens.
func (r *Request) MustPut(url string) *Response {
	resp, err := r.Put(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Put send the request with Put method and specified url.
func (r *Request) Put(url string) (*Response, error) {
	return r.send(http.MethodPut, url)
}

// MustPatch like Patch, panic if error happens.
func (r *Request) MustPatch(url string) *Response {
	resp, err := r.Patch(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Patch send the request with PATCH method and specified url.
func (r *Request) Patch(url string) (*Response, error) {
	return r.send(http.MethodPatch, url)
}

// MustDelete like Delete, panic if error happens.
func (r *Request) MustDelete(url string) *Response {
	resp, err := r.Delete(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Delete send the request with DELETE method and specified url.
func (r *Request) Delete(url string) (*Response, error) {
	return r.send(http.MethodDelete, url)
}

// MustOptions like Options, panic if error happens.
func (r *Request) MustOptions(url string) *Response {
	resp, err := r.Options(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Options send the request with OPTIONS method and specified url.
func (r *Request) Options(url string) (*Response, error) {
	return r.send(http.MethodOptions, url)
}

// MustHead like Head, panic if error happens.
func (r *Request) MustHead(url string) *Response {
	resp, err := r.send(http.MethodHead, url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Head send the request with HEAD method and specified url.
func (r *Request) Head(url string) (*Response, error) {
	return r.send(http.MethodHead, url)
}

// Body set the request body.
func (r *Request) Body(body interface{}) *Request {
	if body == nil {
		return r
	}
	switch b := body.(type) {
	case io.ReadCloser:
		r.httpRequest.Body = b
	case io.Reader:
		r.httpRequest.Body = ioutil.NopCloser(b)
	case []byte:
		r.BodyBytes(b)
	case string:
		r.BodyString(b)
	}
	return r
}

// BodyBytes set the request body as []byte.
func (r *Request) BodyBytes(body []byte) *Request {
	r.httpRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r
}

// BodyString set the request body as string.
func (r *Request) BodyString(body string) *Request {
	r.httpRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r
}

// BodyJsonString set the request body as string and set Content-Type header
// as "application/json; charset=UTF-8"
func (r *Request) BodyJsonString(body string) *Request {
	r.httpRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	r.setContentType(CONTENT_TYPE_APPLICATION_JSON_UTF8)
	return r
}

// BodyJsonBytes set the request body as []byte and set Content-Type header
// as "application/json; charset=UTF-8"
func (r *Request) BodyJsonBytes(body []byte) *Request {
	r.httpRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	r.setContentType(CONTENT_TYPE_APPLICATION_JSON_UTF8)
	return r
}

// BodyJsonMarshal set the request body that marshaled from object, and
// set Content-Type header as "application/json; charset=UTF-8"
func (r *Request) BodyJsonMarshal(v interface{}) *Request {
	b, err := json.Marshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.BodyBytes(b)
}

func (r *Request) setContentType(contentType string) *Request {
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

// Send sends the request.
func (r *Request) Send() (resp *Response, err error) {
	return r.execute()
}
