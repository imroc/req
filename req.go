package req

import (
	"bytes"
	"crypto/tls"
	"fmt"
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
	Resp      Response
	done      bool
	body      []byte
	Client    http.Client
}

// InsecureTLS insecure the https.
func (r *Request) InsecureTLS() *Request {
	r.Client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return r
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
		r.body = []byte(v)
	case []byte:
		bf := bytes.NewBuffer(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
		r.body = v
	}
	return r
}

func (r *Request) GetBody() []byte {
	return r.body
}

// Bytes execute the request and get the response body as []byte.
func (r *Request) Bytes() (data []byte, err error) {
	resp, err := r.Response()
	if err != nil {
		return
	}
	data = resp.Body
	return
}

// MustBytes execute the request and get the response body as []byte.panic if error happens.
func (r *Request) MustBytes() (data []byte) {
	resp, err := r.Response()
	if err != nil {
		panic(err)
	}
	data = resp.Body
	return
}

// String execute the request and get the response body as string.
func (r *Request) String() (s string, err error) {
	resp, err := r.Response()
	if err != nil {
		return
	}
	s = string(resp.Body)
	return
}

// MustString execute the request and get the response body as string.panic if error happens.
func (r *Request) MustString() (s string) {
	resp, err := r.Response()
	if err != nil {
		panic(err)
	}
	s = string(resp.Body)
	return
}

// String execute the request and get the response body unmarshal to json.
func (r *Request) ToJson(v interface{}) (err error) {
	resp, err := r.Response()
	if err != nil {
		return
	}
	err = resp.ToJson(v)
	return
}

// String execute the request and get the response body unmarshal to xml.
func (r *Request) ToXml(v interface{}) (err error) {
	resp, err := r.Response()
	if err != nil {
		return
	}
	err = resp.ToXml(v)
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
func (r *Request) Response() (resp Response, err error) {
	if r.Resp.Raw != nil { // provent multiple call
		resp = r.Resp
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
	r.Resp.Raw, err = r.Client.Do(r.req)
	if err != nil {
		return
	}
	err = r.Resp.Receive()
	if err != nil {
		return
	}
	resp = r.Resp
	r.done = true
	return
}

// MustResponse execute the request and get the response.panic if error happens.
func (r *Request) MustResponse(resp Response) {
	resp, err := r.Response()
	if err != nil {
		panic(err)
	}
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
func Wrap(url string, req *http.Request) *Request {
	return &Request{
		url:       url,
		urlEncode: true,
		params:    M{},
		req:       req,
		Resp:      Response{},
	}
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
		Resp: Response{},
	}
}

func (r *Request) Format(s fmt.State, verb rune) {

	if s.Flag('+') { // include header and format pretty.
		fmt.Fprint(s, r.req.Method, " ", r.Url(), " ", r.req.Proto)
		for name, values := range r.req.Header {
			for _, value := range values {
				fmt.Fprint(s, "\n", name, ":", value)
			}
		}
		if r.body != nil {
			fmt.Fprint(s, "\n\n", string(r.body))
		}
		if r.Resp.Raw != nil {
			fmt.Fprint(s, "\n\n")
			r.Resp.Format(s, verb)
		}
		return
	}

	fmt.Fprint(s, r.req.Method, " ", r.Url())

	pretty := false
	if (r.body != nil && bytes.IndexByte(r.body, '\n') != -1) || (r.Resp.Body != nil && bytes.IndexByte(r.Resp.Body, '\n') != -1) {
		pretty = true
	}
	if pretty {
		fmt.Fprint(s, "\n")
		if r.body != nil {
			fmt.Fprint(s, string(r.body))
		}
		if r.Resp.Body != nil {
			fmt.Fprint(s, "\n", string(r.Resp.Body))
		}
		return
	}
	if r.body != nil {
		fmt.Fprint(s, " ", string(r.body))
	}
	if r.Resp.Body != nil { // request has been sent.
		fmt.Fprint(s, " ", string(r.Resp.Body))
	}
}
