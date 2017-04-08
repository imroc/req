package req

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var regBlank = regexp.MustCompile(`\s+`)

// M represents the request params or headers.
type M map[string]string
type formFile struct {
	filename string
	file     io.Reader
}

// Request provides much easier usage than http.Request
type Request struct {
	err     error
	url     string
	params  M
	files   map[string]formFile
	req     *http.Request
	resp    *Response
	body    []byte
	setting *setting
}

var ErrNilReqeust = errors.New("nil request")

// GetRequest return the raw *http.Request inside the Request.
func (r *Request) GetRequest() *http.Request {
	if r == nil {
		return nil
	}
	return r.req
}

// Proto set the protocol version to the request.
func (r *Request) Proto(vers string) *Request {
	if r == nil {
		return nil
	}
	if r.req == nil {
		r.req = basicRequest()
	}

	if len(vers) == 0 {
		vers = "HTTP/1.1"
	}
	major, minor, ok := http.ParseHTTPVersion(vers)
	if ok {
		r.req.Proto = vers
		r.req.ProtoMajor = major
		r.req.ProtoMinor = minor
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

// File upload the file with specified form name and file name,
// and the file is come from disk with the given file name(relative path or absolute path).
func (r *Request) File(formname, name string) *Request {
	if r == nil {
		return nil
	}
	if r.files == nil {
		r.files = make(map[string]formFile)
	}
	file, err := os.Open(name)
	if err != nil {
		r.err = err
		return r
	}
	r.files[formname] = formFile{filepath.Base(name), file}
	return r
}

// FileReader upload the file with specified form name and file name,
// and the file is come from io.Reader.
func (r *Request) FileReader(formname, filename string, file io.Reader) *Request {
	if r == nil {
		return nil
	}
	if r.files == nil {
		r.files = make(map[string]formFile)
	}
	r.files[formname] = formFile{filename, file}
	return r
}

// Cookie set the cookie for to the request.
func (r *Request) Cookie(cookie *http.Cookie) *Request {
	if r == nil {
		return nil
	}
	if r.req == nil {
		r.req = basicRequest()
	}
	r.req.Header.Add("Cookie", cookie.String())
	return r
}

// BasicAuth sets the request's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (r *Request) BasicAuth(username, password string) *Request {
	if r == nil {
		return nil
	}
	if r.req == nil {
		r.req = basicRequest()
	}
	r.req.SetBasicAuth(username, password)
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

// BodyJSON set the request body as json, support string, []byte or pointer of struct which has json tag.
// it set the Content-Type header to application/json.
func (r *Request) BodyJSON(body interface{}) *Request {
	if r == nil || r.req == nil {
		return nil
	}
	switch v := body.(type) {
	case string:
		bf := bytes.NewBufferString(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
		r.req.Header.Set("Content-Type", "application/json")
		r.body = []byte(v)
	case []byte:
		bf := bytes.NewBuffer(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
		r.req.Header.Set("Content-Type", "application/json")
		r.body = v
	default:
		bs, err := json.Marshal(body)
		if err == nil {
			bf := bytes.NewBuffer(bs)
			r.req.Body = ioutil.NopCloser(bf)
			r.req.ContentLength = int64(len(bs))
			r.req.Header.Set("Content-Type", "application/json")
			r.body = bs
		}
	}
	return r
}

// BodyXML set the request body as xml, support string, []byte or pointer of struct which has xml tag.
// it set the Content-Type header to text/xml
func (r *Request) BodyXML(body interface{}) *Request {
	if r == nil || r.req == nil {
		return nil
	}
	switch v := body.(type) {
	case string:
		bf := bytes.NewBufferString(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
		r.req.Header.Set("Content-Type", "text/xml")
		r.body = []byte(v)
	case []byte:
		bf := bytes.NewBuffer(v)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(v))
		r.req.Header.Set("Content-Type", "text/xml")
		r.body = v
	default:
		bs, err := xml.Marshal(body)
		if err == nil {
			bf := bytes.NewBuffer(bs)
			r.req.Body = ioutil.NopCloser(bf)
			r.req.ContentLength = int64(len(bs))
			r.req.Header.Set("Content-Type", "text/xml")
			r.body = bs
		}
	}
	return r
}

// GetBody return the request body.
func (r *Request) GetBody() []byte {
	if r == nil {
		return nil
	}
	if r.body == nil && r.req != nil && (r.req.Method == "POST" || r.req.Method == "PUT") {
		return []byte(r.getParamBody())
	}
	return r.body
}

// ReceiveBytes execute the request and get the response body as []byte.
// err is not nil if error happens during the reqeust been executed.
func (r *Request) ReceiveBytes() (data []byte, err error) {
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

// ToJSON execute the request and get the response body unmarshal to json.
func (r *Request) ToJSON(v interface{}) (err error) {
	resp, err := r.ReceiveResponse()
	if err != nil {
		return
	}
	err = resp.ToJSON(v)
	return
}

// ToXML execute the request and get the response body unmarshal to xml.
func (r *Request) ToXML(v interface{}) (err error) {
	resp, err := r.ReceiveResponse()
	if err != nil {
		return
	}
	err = resp.ToXML(v)
	return
}

func (r *Request) getParamBody() string {
	if len(r.params) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for k, v := range r.params {
		k = url.QueryEscape(k)
		v = url.QueryEscape(v)
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
	r.Header("Content-Type", "application/x-www-form-urlencoded")
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
func (r *Request) Url(url string) *Request {
	if r == nil {
		return nil
	}
	r.url = url
	return r
}

// Host set the request's Host.
func (r *Request) Host(host string) *Request {
	if r == nil || r.req == nil {
		return nil
	}
	r.req.Host = host
	return r
}

// ReceiveResponse execute the request and get the response body as *Response,
// err is not nil if error happens during the reqeust been executed.
func (r *Request) ReceiveResponse() (resp *Response, err error) {
	if r == nil {
		err = ErrNilReqeust
		return
	}
	if r.err != nil {
		err = r.err
	}
	if r.resp != nil { // provent multiple call
		resp = r.resp
		return
	}
	err = r.Do()
	if err != nil {
		r.err = err
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
	r.err = nil
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

	switch r.req.Method {
	case "GET":
		if len(r.params) > 0 {
			destUrl = r.buildGetUrl()
		}
	case "POST", "PUT":
		if len(r.files) > 0 { // upload file
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formname, form := range r.files {
					fileWriter, err := bodyWriter.CreateFormFile(formname, form.filename)
					if err != nil {
						r.err = err
						return
					}
					//iocopy
					_, err = io.Copy(fileWriter, form.file)
					if closer, ok := form.file.(io.Closer); ok {
						closer.Close()
					}
					if err != nil {
						r.err = err
						return
					}
				}
				for k, v := range r.params {
					bodyWriter.WriteField(k, v)
				}
				bodyWriter.Close()
				pw.Close()
			}()
			r.Header("Content-Type", bodyWriter.FormDataContentType())
			r.req.Body = ioutil.NopCloser(pr)
		} else { // set params to body
			if len(r.params) > 0 {
				r.Header("Content-Type", "application/x-www-form-urlencoded")
				r.Body(r.getParamBody())
			}
		}
	}

	// set url
	u, err := url.Parse(destUrl)
	if err != nil {
		return
	}
	r.req.URL = u
	respRaw, err := r.GetClient().Do(r.req)
	if err != nil {
		return
	}
	resp := WrapResponse(respRaw)
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
	resp, _ = r.ReceiveResponse()
	return
}

// Method set the method for the request.
func (r *Request) Method(method string) *Request {
	if r == nil {
		return nil
	}
	if r.req == nil {
		r.req = basicRequest()
		r.req.Method = method
	} else {
		r.req.Method = method
	}
	return r
}

// Format implements fmt.Formatter, format the request's information.
func (r *Request) Format(s fmt.State, verb rune) {
	if r == nil || r.req == nil {
		return
	}
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
	} else if s.Flag('-') { // keep all informations in one line.
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

// Get create a new  *Request with GET method.
func Get(url string) *Request {
	return newRequest(url, "GET")
}

// Post create a new  *Request with POST method.
func Post(url string) *Request {
	return newRequest(url, "POST")
}

// Put create a new  *Request with PUT method.
func Put(url string) *Request {
	return newRequest(url, "PUT")
}

// Delete create a new  *Request with DELETE method.
func Delete(url string) *Request {
	return newRequest(url, "DELETE")
}

// Head create a new  *Request with HEAD method.
func Head(url string) *Request {
	return newRequest(url, "HEAD")
}

// New create a new Request with the underlying *http.Request.
func New() *Request {
	return &Request{
		params: M{},
		req:    basicRequest(),
	}
}

// WrapRequest wraps the *http.Request to the *req.Request.
func WrapRequest(req *http.Request) *Request {
	return &Request{
		params: M{},
		req:    req,
	}
}

func newRequest(url, method string) *Request {
	req := basicRequest()
	req.Method = method
	return &Request{
		url:    url,
		params: M{},
		req:    req,
	}
}

func basicRequest() *http.Request {
	return &http.Request{
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
}
