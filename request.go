package req

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/go-multierror"
	"github.com/imroc/req/v2/internal/util"
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
	Headers        http.Header
	Result         interface{}
	Error          interface{}
	error          error
	client         *Client
	RawRequest     *http.Request
	isSaveResponse bool
	isMultiPart    bool
	output         io.WriteCloser
}

// New create a new request using the global default client.
func New() *Request {
	return defaultClient.R()
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

func (r *Request) SetResult(result interface{}) *Request {
	r.Result = util.GetPointer(result)
	return r
}

func (r *Request) SetError(error interface{}) *Request {
	r.Error = util.GetPointer(error)
	return r
}

func (r *Request) SetHeaders(hdrs map[string]string) *Request {
	for k, v := range hdrs {
		r.SetHeader(k, v)
	}
	return r
}

// SetHeader set the common header for all requests.
func (r *Request) SetHeader(key, value string) *Request {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set(key, value)
	return r
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

func (r *Request) Send(method, url string) (*Response, error) {
	if r.error != nil {
		return nil, r.error
	}
	r.RawRequest.Method = method
	r.URL = url
	return r.client.Do(r)
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

// SetBodyBytes set the request body as []byte.
func (r *Request) SetBodyBytes(body []byte) *Request {
	r.RawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r
}

// SetBodyString set the request body as string.
func (r *Request) SetBodyString(body string) *Request {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r
}

// SetBodyJsonString set the request body as string and set Content-Type header
// as "application/json; charset=UTF-8"
func (r *Request) SetBodyJsonString(body string) *Request {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	r.SetContentType(CONTENT_TYPE_APPLICATION_JSON_UTF8)
	return r
}

// SetBodyJsonBytes set the request body as []byte and set Content-Type header
// as "application/json; charset=UTF-8"
func (r *Request) SetBodyJsonBytes(body []byte) *Request {
	r.RawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
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
	r.RawRequest.Header.Set("Content-Type", contentType)
	return r
}
