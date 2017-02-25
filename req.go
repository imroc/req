package req

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// P represents the request params.
type P map[string]string

// Request provides much easier useage than http.Request
type Request struct {
	url    string
	params P
	req    *http.Request
	resp   *http.Response
	body   []byte
}

// Param set single param to the request.
func (r *Request) Param(k, v string) *Request {
	r.params[k] = v
	return r
}

// Params set multiple params to the request.
func (r *Request) Params(params P) *Request {
	for k, v := range params {
		r.params[k] = v
	}
	return r
}

// Header set the request header.
func (r *Request) Header(k, v string) *Request {
	r.req.Header.Set(k, v)
	return r
}

// Body set the request body,support string and []byte.
func (r *Request) Body(body interface{}) *Request {
	switch v := body.(type) {
	case string:
		bf := bytes.NewBufferString(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
	case []byte:
		bf := bytes.NewBuffer(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
	}
	return r
}

// Bytes execute the request and get the response body as []byte.
func (r *Request) Bytes() (data []byte, err error) {
	if r.body != nil { // in case multiple call
		data = r.body
		return
	}
	resp, err := r.Response()
	if err != nil {
		return
	}
	defer resp.Body.Close()
	r.body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	data = r.body
	return
}

// String execute the request and get the response body as string.
func (r *Request) String() (s string, err error) {
	data, err := r.Bytes()
	if err != nil {
		return
	}
	s = string(data)
	return
}

// String execute the request and get the response body unmarshal to json.
func (r *Request) ToJson(v interface{}) (err error) {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// String execute the request and get the response body unmarshal to xml.
func (r *Request) ToXml(v interface{}) (err error) {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// Response execute the request and get the response.
func (r *Request) Response() (resp *http.Response, err error) {
	if r.resp != nil { // provent multiple call
		resp = r.resp
		return
	}
	// handle request params
	if len(r.params) > 0 {
		var buf bytes.Buffer
		for k, v := range r.params {
			buf.WriteString(url.QueryEscape(k))
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
			buf.WriteByte('&')
		}
		p := buf.String()
		p = p[0 : len(p)-1]
		switch r.req.Method {
		case "GET":
			if strings.Index(r.url, "?") != -1 {
				r.url += "&" + p
			} else {
				r.url += "?" + p
			}
		case "POST":
			if r.req.Body == nil {
				r.Header("Content-Type", "application/x-www-form-urlencoded")
				r.Body(p)
			}
		}
	}
	// set url
	u, err := url.Parse(r.url)
	if err != nil {
		return
	}
	r.req.URL = u
	resp, err = http.DefaultClient.Do(r.req)
	return
}

// Get returns *Request with GET method.
func Get(url string) *Request {
	return newRequest(url, "GET")
}

// Get returns *Request with POST method.
func Post(url string) *Request {
	return newRequest(url, "POST")
}

func newRequest(url, method string) *Request {
	return &Request{
		url:    url,
		params: P{},
		req: &http.Request{
			Method:     method,
			Header:     make(http.Header),
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		},
	}
}
