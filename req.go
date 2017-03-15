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

// M represents the request params.
type M map[string]string

// Request provides much easier useage than http.Request
type Request struct {
	url       string
	urlEncode bool
	params    M
	req       *http.Request
	resp      *http.Response
	done      bool
	respBody  []byte
	reqBody   []byte
}

// Param set single param to the request.
func (r *Request) Param(k, v string) *Request {
	r.params[k] = v
	return r
}

// Params set multiple params to the request.
func (r *Request) Params(params M) *Request {
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

func (r *Request) Headers(params M) *Request {
	for k, v := range params {
		r.req.Header.Set(k, v)
	}
	return r
}

// Body set the request body,support string and []byte.
func (r *Request) Body(body interface{}) *Request {
	switch v := body.(type) {
	case string:
		bf := bytes.NewBufferString(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
		r.reqBody = []byte(v)
	case []byte:
		bf := bytes.NewBuffer(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
		r.reqBody = v
	}
	return r
}

func (r *Request) BodyBytes() []byte {
	return r.reqBody
}

func (r *Request) BodyString() string {
	return string(r.reqBody)
}

// Bytes execute the request and get the response body as []byte.
func (r *Request) Bytes() (data []byte, err error) {
	if r.respBody != nil { // in case multiple call
		data = r.respBody
		return
	}
	resp, err := r.Response()
	if err != nil {
		return
	}
	defer resp.Body.Close()
	r.respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	data = r.respBody
	return
}

func (r *Request) MustBytes() []byte {
	data, err := r.Bytes()
	if err != nil {
		panic(err)
	}
	return data
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

func (r *Request) MustString() string {
	s, err := r.String()
	if err != nil {
		panic(err)
	}
	return s
}

// String execute the request and get the response body unmarshal to json.
func (r *Request) ToJson(v interface{}) (err error) {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	return
}

// String execute the request and get the response body unmarshal to xml.
func (r *Request) ToXml(v interface{}) (err error) {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	err = xml.Unmarshal(data, v)
	return
}

func (r *Request) UrlEncode(urlEncode bool) *Request {
	r.urlEncode = urlEncode
	return r
}

func (r *Request) getParamBody() string {
	var buf bytes.Buffer
	for k, v := range r.params {
		if r.urlEncode {
			k = url.QueryEscape(k)
			v = url.QueryEscape(v)
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
		buf.WriteByte('&')
	}
	p := buf.String()
	p = p[0 : len(p)-1]
	return p
}

func (r *Request) buildGetUrl() string {
	ret := r.url
	p := r.getParamBody()
	if strings.Index(r.url, "?") != -1 {
		ret += "&" + p
	} else {
		ret += "?" + p
	}
	return ret
}

func (r *Request) setParamBody() {
	if r.urlEncode {
		r.Header("Content-Type", "application/x-www-form-urlencoded")
	}
	r.Body(r.getParamBody())
}

func (r *Request) Url() string {
	if r.req.Method != "GET" || r.done {
		return r.url
	}
	return r.buildGetUrl() //GET method and did not send request yet.
}

// Response execute the request and get the response.
func (r *Request) Response() (resp *http.Response, err error) {
	if r.resp != nil { // provent multiple call
		resp = r.resp
		return
	}
	// handle request params
	if len(r.params) > 0 {
		switch r.req.Method {
		case "GET":
			r.url = r.buildGetUrl()
		case "POST":
			if r.req.Body == nil {
				r.setParamBody()
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
	r.done = true
	return
}

func (r *Request) MustResponse() *http.Response {
	resp, err := r.Response()
	if err != nil {
		panic(err)
	}
	return resp
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
		url:       url,
		urlEncode: true,
		params:    M{},
		req: &http.Request{
			Method:     method,
			Header:     make(http.Header),
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		},
	}
}
