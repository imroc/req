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

type Request struct {
	error       error
	client      *Client
	httpRequest *http.Request
}

func New() *Request {
	return defaultClient.R()
}

func (r *Request) appendError(err error) {
	r.error = multierror.Append(r.error, err)
}

func (r *Request) Error() error {
	return r.error
}

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

func (r *Request) MustGet(url string) *Response {
	resp, err := r.Get(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func (r *Request) Get(url string) (*Response, error) {
	return r.send(http.MethodGet, url)
}

func (r *Request) MustPost(url string) *Response {
	resp, err := r.Post(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func (r *Request) Post(url string) (*Response, error) {
	return r.send(http.MethodPost, url)
}

func (r *Request) MustPut(url string) *Response {
	resp, err := r.Put(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func (r *Request) Put(url string) (*Response, error) {
	return r.send(http.MethodPut, url)
}

func (r *Request) MustPatch(url string) *Response {
	resp, err := r.Patch(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func (r *Request) Patch(url string) (*Response, error) {
	return r.send(http.MethodPatch, url)
}

func (r *Request) MustDelete(url string) *Response {
	resp, err := r.Delete(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func (r *Request) Delete(url string) (*Response, error) {
	return r.send(http.MethodDelete, url)
}

func (r *Request) MustOptions(url string) *Response {
	resp, err := r.Options(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func (r *Request) Options(url string) (*Response, error) {
	return r.send(http.MethodOptions, url)
}

func (r *Request) MustHead(url string) (*Response, error) {
	return r.send(http.MethodHead, url)
}

func (r *Request) Head(url string) (*Response, error) {
	return r.send(http.MethodHead, url)
}

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

func (r *Request) BodyBytes(body []byte) *Request {
	r.httpRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return r
}

func (r *Request) BodyString(body string) *Request {
	r.httpRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	return r
}

func (r *Request) BodyJsonString(body string) *Request {
	r.httpRequest.Body = ioutil.NopCloser(strings.NewReader(body))
	r.setContentType(CONTENT_TYPE_APPLICATION_JSON_UTF8)
	return r
}

func (r *Request) BodyJsonBytes(body []byte) *Request {
	r.httpRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	r.setContentType(CONTENT_TYPE_APPLICATION_JSON_UTF8)
	return r
}

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
	httpResponse, err := r.client.httpClient.Do(r.httpRequest)
	if err != nil {
		return
	}
	resp = &Response{
		request:  r,
		Response: httpResponse,
	}
	if r.client.t.AutoDiscard {
		err = resp.Discard()
	}
	return
}

func (r *Request) Send() (resp *Response, err error) {
	return r.execute()
}
