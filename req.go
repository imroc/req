package req

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var regBlank = regexp.MustCompile(`\s+`)

// M represents the request params.
type M map[string]string

// Request provides much easier useage than http.Request
type Request struct {
	url       string
	urlEncode bool
	params    M
	req       *http.Request
	resp      *Response
	body      []byte
	client    http.Client
}

var ErrNilReqeust = errors.New("nil request")

// Request return the raw *http.Request inside the Request.
func (r *Request) Request() *http.Request {
	if r == nil {
		return nil
	}
	return r.req
}

// InsecureTLS insecure the https.
func (r *Request) InsecureTLS() *Request {
	if r == nil {
		return nil
	}
	r.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return r
}

// Param set one single param to the request.
func (r *Request) Param(k, v string) *Request {
	if r == nil {
		return nil
	}
	r.params[k] = v
	return r
}

// Params set multiple params to the request.
func (r *Request) Params(params M) *Request {
	if r == nil {
		return nil
	}
	for k, v := range params {
		r.params[k] = v
	}
	return r
}

// Header set one single header to the  request.
func (r *Request) Header(k, v string) *Request {
	if r == nil || r.req == nil {
		return nil
	}
	r.req.Header.Set(k, v)
	return r
}

// Headers set multiple headers to the request.
func (r *Request) Headers(params M) *Request {
	if r == nil || r.req == nil {
		return nil
	}
	for k, v := range params {
		r.req.Header.Set(k, v)
	}
	return r
}

// Body set the request body,support string and []byte.
func (r *Request) Body(body interface{}) *Request {
	if r == nil || r.req == nil {
		return nil
	}
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

// GetBody return the request body.
func (r *Request) GetBody() []byte {
	if r == nil {
		return nil
	}
	if r.body == nil && r.req != nil && r.req.Method == "POST" {
		return []byte(r.getParamBody())
	}
	return r.body
}

// ReceiveBytes execute the request and get the response body as []byte.
// err is not nil if error happens during the reqeust been executed.
func (r *Request) ReceiveBytes() (data []byte, err error) {
	if r == nil {
		err = ErrNilReqeust
		return
	}
	resp, err := r.ReceiveResponse()
	if err != nil {
		return
	}
	data, err = resp.ReceiveBytes()
	return
}

// Bytes execute the request and get the response body as []byte.
// data is nil if error happens if error happens during the reqeust been executed.
func (r *Request) Bytes() (data []byte) {
	data, _ = r.ReceiveBytes()
	return
}

// ReceiveString execute the request and get the response body as string.
// err is not nil if error happens during the reqeust been executed.
func (r *Request) ReceiveString() (s string, err error) {
	if r == nil {
		err = ErrNilReqeust
		return
	}
	resp, err := r.ReceiveResponse()
	if err != nil {
		return
	}
	s, err = resp.ReceiveString()
	return
}

// String execute the request and get the response body as string,
// s equals "" if error happens.
func (r *Request) String() (s string) {
	s, _ = r.ReceiveString()
	return
}

// ToJson execute the request and get the response body unmarshal to json.
func (r *Request) ToJson(v interface{}) (err error) {
	if r == nil {
		err = ErrNilReqeust
		return
	}
	resp, err := r.ReceiveResponse()
	if err != nil {
		return
	}
	err = resp.ToJson(v)
	return
}

// ToXml execute the request and get the response body unmarshal to xml.
func (r *Request) ToXml(v interface{}) (err error) {
	if r == nil {
		err = ErrNilReqeust
		return
	}
	resp, err := r.ReceiveResponse()
	if err != nil {
		return
	}
	err = resp.ToXml(v)
	return
}

// UrlEncode set weighter the params should be url encoded or not, default to true.
func (r *Request) UrlEncode(urlEncode bool) *Request {
	if r == nil {
		return nil
	}
	r.urlEncode = urlEncode
	return r
}

func (r *Request) getParamBody() string {
	if len(r.params) == 0 {
		return ""
	}
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
	if p := r.getParamBody(); p != "" {
		if strings.Index(r.url, "?") != -1 {
			ret += "&" + p
		} else {
			ret += "?" + p
		}
	}
	return ret
}

func (r *Request) setParamBody() {
	if r.urlEncode {
		r.Header("Content-Type", "application/x-www-form-urlencoded")
	}
	r.Body(r.getParamBody())
}

// GetUrl return the url of the request.
func (r *Request) GetUrl() string {
	if r == nil {
		return ""
	}
	if r.req != nil && r.req.Method == "GET" {
		return r.buildGetUrl() //GET method and did not send request yet.
	}
	return r.url
}

// Url set the request's url.
func (r *Request) Url(urlStr string) *Request {
	if r == nil {
		return nil
	}
	r.url = urlStr
	return r
}

// ReceiveResponse execute the request and get the response body as *Response,
// err is not nil if error happens during the reqeust been executed.
func (r *Request) ReceiveResponse() (resp *Response, err error) {
	if r == nil {
		err = ErrNilReqeust
		return
	}
	if r.resp != nil { // provent multiple call
		resp = r.resp
		return
	}
	err = r.Do()
	if err != nil {
		return
	}
	resp = r.resp
	return
}

// Undo let the request could be executed again. The request will only been
// executed only once by default, if you want to reuse the request, making it
// could be executed again, even change some params or headers, you can Undo
// the request and try again..
func (r *Request) Undo() *Request {
	if r == nil {
		return nil
	}
	r.resp = nil
	return r
}

// Do execute the request. return error if error happens. note, it will always
// execute the request. even it has been executed before.
func (r *Request) Do() (err error) {
	if r == nil {
		err = ErrNilReqeust
	}
	// handle request params
	destUrl := r.url
	if len(r.params) > 0 {
		switch r.req.Method {
		case "GET":
			destUrl = r.buildGetUrl()
		case "POST":
			r.setParamBody()
		}
	}
	// set url
	u, err := url.Parse(destUrl)
	if err != nil {
		return
	}
	r.req.URL = u
	respRaw, err := r.client.Do(r.req)
	if err != nil {
		return
	}
	resp := NewResponse(respRaw)
	err = resp.Receive()
	if err != nil {
		return
	}
	r.resp = resp
	return
}

// Response execute the request and get the response body as *Response,
// resp equals nil if error happens.
func (r *Request) Response() (resp *Response) {
	if r == nil {
		return nil
	}
	resp, _ = r.ReceiveResponse()
	return
}

// Get create a new  *Request with GET method.
func Get(url string) *Request {
	return newRequest(url, "GET")
}

// Post create a new  *Request with POST method.
func Post(url string) *Request {
	return newRequest(url, "POST")
}

// New create a new Request with the underlying *http.Request.
func New(req *http.Request) *Request {
	return &Request{
		urlEncode: true,
		params:    M{},
		req:       req,
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
	}
}

// Format implements fmt.Formatter, format the request's infomation.
func (r *Request) Format(s fmt.State, verb rune) {
	if s.Flag('+') { // include header and format pretty.
		fmt.Fprint(s, r.req.Method, " ", r.GetUrl(), " ", r.req.Proto)
		var resp *Response
		if verb != 'r' {
			resp = r.Response()
		}
		for name, values := range r.req.Header {
			for _, value := range values {
				fmt.Fprint(s, "\n", name, ":", value)
			}
		}
		if len(r.body) > 0 {
			fmt.Fprint(s, "\n\n", string(r.body))
		}
		if resp != nil {
			fmt.Fprint(s, "\n\n")
			resp.Format(s, verb)
		}
	} else if s.Flag('-') { // keep all infomations in one line.
		fmt.Fprint(s, r.req.Method, " ", r.GetUrl())
		if len(r.body) > 0 {
			str := regBlank.ReplaceAllString(string(r.body), "")
			fmt.Fprint(s, str)
		}
		if str := r.String(); str != "" {
			str = regBlank.ReplaceAllString(str, "")
			fmt.Fprint(s, " ", str)
		}
	} else { // auto
		fmt.Fprint(s, r.req.Method, " ", r.GetUrl())
		if verb == 'r' {
			if len(r.body) > 0 {
				if bytes.IndexByte(r.body, '\n') != -1 && r.body[0] != '\n' {
					fmt.Fprint(s, "\n")
				}
				fmt.Fprint(s, string(r.body))
			}
		} else {
			respBody := r.Bytes()
			if (len(r.body) > 0 && bytes.IndexByte(r.body, '\n') != -1) || (len(respBody) > 0 && bytes.IndexByte(respBody, '\n') != -1) { // pretty format
				if len(r.body) > 0 {
					fmt.Fprint(s, "\n", string(r.body))
				}
				if len(respBody) > 0 {
					fmt.Fprint(s, "\n", string(respBody))
				}
			} else {
				if len(r.body) > 0 {
					fmt.Fprint(s, " ", string(r.body))
				}
				if len(respBody) > 0 {
					fmt.Fprint(s, " ", string(respBody))
				}
			}
		}
	}

}
