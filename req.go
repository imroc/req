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
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// http request header param
type Header map[string]string

// http request param
type Param map[string]string

// represents a file to upload
type FileUpload struct {
	// filename in multipart form.
	FileName string
	// form field name
	FieldName string
	// file to uplaod, required
	File io.ReadCloser
}

// Debug enable debug mode if set to true
var Debug bool

// ShowCost show the time spent by the request if set to true
var ShowCost bool

var defaultClient *http.Client
var regTextContentType = regexp.MustCompile("xml|json|text")

func init() {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	defaultClient = &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   2 * time.Minute,
	}
}

// Req represents a request with it's response
type Req struct {
	req      *http.Request
	resp     *http.Response
	client   *http.Client
	reqBody  []byte
	respBody []byte
	cost     time.Duration
}

func (r *Req) getReqBody() io.ReadCloser {
	if r.reqBody == nil {
		return nil
	}
	return ioutil.NopCloser(bytes.NewReader(r.reqBody))
}

// Do execute request.
func Do(method, rawurl string, v ...interface{}) (r *Req, err error) {
	if rawurl == "" {
		return nil, errors.New("req: url not specified")
	}
	req := &http.Request{
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	r = &Req{req: req}
	handleBody := func(b *body) {
		if b == nil {
			return
		}
		r.reqBody = b.Data
		req.Body = r.getReqBody()
		req.ContentLength = int64(len(r.reqBody))
		if b.ContentType != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", b.ContentType)
		}
	}

	var param []Param
	var file []FileUpload
	for _, p := range v {
		switch t := p.(type) {
		case Header:
			for key, value := range t {
				req.Header.Add(key, value)
			}
		case http.Header:
			req.Header = t
		case io.Reader:
			bs, err := ioutil.ReadAll(t)
			if err != nil {
				return nil, err
			}
			handleBody(&body{Data: bs})
			if rc, ok := t.(io.ReadCloser); ok {
				rc.Close()
			}
		case *body:
			handleBody(t)
		case Param:
			param = append(param, t)
		case string:
			handleBody(&body{Data: []byte(t)})
		case []byte:
			handleBody(&body{Data: []byte(t)})
		case *http.Client:
			r.client = t
		case FileUpload:
			file = append(file, t)
		case []FileUpload:
			if file == nil {
				file = make([]FileUpload, 0)
			}
			file = append(file, t...)
		case *http.Cookie:
			req.AddCookie(t)
		case error:
			err = t
			return
		}
	}

	if len(file) > 0 && (req.Method == "POST" || req.Method == "PUT") {
		pr, pw := io.Pipe()
		bodyWriter := multipart.NewWriter(pw)
		go func() {
			for _, p := range param {
				for key, value := range p {
					bodyWriter.WriteField(key, value)
				}
			}
			i := 0
			for _, f := range file {
				if f.FieldName == "" {
					i++
					f.FieldName = "file" + strconv.Itoa(i)
				}
				fileWriter, e := bodyWriter.CreateFormFile(f.FieldName, f.FileName)
				if e != nil {
					err = e
					return
				}
				//iocopy
				_, e = io.Copy(fileWriter, f.File)
				if e != nil {
					err = e
					return
				}
				f.File.Close()
			}
			bodyWriter.Close()
			pw.Close()
		}()
		req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
		req.Body = ioutil.NopCloser(pr)
	} else if len(param) > 0 {
		params := make(url.Values)
		for _, p := range param {
			for key, value := range p {
				params.Add(key, value)
			}
		}
		paramStr := params.Encode()
		if method == "GET" {
			if strings.IndexByte(rawurl, '?') == -1 {
				rawurl = rawurl + "?" + paramStr
			} else {
				rawurl = rawurl + "&" + paramStr
			}
		} else {
			if req.Body != nil {
				return nil, errors.New("req: can not set both body and params")
			}
			body := &body{
				ContentType: "application/x-www-form-urlencoded",
				Data:        []byte(paramStr),
			}
			handleBody(body)
		}
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req.URL = u

	if r.client == nil {
		r.client = defaultClient
	}

	now := time.Now()
	resp, errDo := r.client.Do(req)
	r.cost = time.Since(now)
	if err != nil {
		return r, err
	}
	if errDo != nil {
		return r, errDo
	}
	r.resp = resp
	ct := resp.Header.Get("Content-Type")
	if ct == "" || regTextContentType.MatchString(ct) {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return r, err
		}
		r.respBody = respBody
	}
	if Debug {
		fmt.Println(r.dump())
	}
	return
}

// Cost returns time spent by the request
func (r *Req) Cost() time.Duration {
	return r.cost
}

// Body represents request's body
type body struct {
	ContentType string
	Data        []byte
}

// BodyXML get request's body as xml
func BodyXML(v interface{}) interface{} {
	b := new(body)
	switch t := v.(type) {
	case string:
		b.Data = []byte(t)
	case []byte:
		b.Data = t
	default:
		bs, err := xml.Marshal(v)
		if err != nil {
			return err
		}
		b.Data = bs
	}
	b.ContentType = "text/xml"
	return b
}

// BodyJSON get request's body as json
func BodyJSON(v interface{}) interface{} {
	b := new(body)
	switch t := v.(type) {
	case string:
		b.Data = []byte(t)
	case []byte:
		b.Data = t
	default:
		bs, err := json.Marshal(v)
		if err != nil {
			return err
		}
		b.Data = bs
	}
	b.ContentType = "text/json"
	return b
}

// File upload files matching the name pattern such as
// /usr/*/bin/go* (assuming the Separator is '/')
func File(patterns ...string) interface{} {
	matches := []string{}
	for _, pattern := range patterns {
		m, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		matches = append(matches, m...)
	}
	if len(matches) == 0 {
		return errors.New("req: No file have been matched")
	}
	uploads := []FileUpload{}
	for _, match := range matches {
		if s, e := os.Stat(match); e != nil || s.IsDir() {
			continue
		}
		file, _ := os.Open(match)
		uploads = append(uploads, FileUpload{File: file, FileName: filepath.Base(match)})
	}

	return uploads
}

// Request returns *http.Request
func (r *Req) Request() *http.Request {
	return r.req
}

// Response returns *http.Response
func (r *Req) Response() *http.Response {
	return r.resp
}

// Bytes returns response body as []byte
func (r *Req) Bytes() []byte {
	return r.respBody
}

// String returns response body as string
func (r *Req) String() string {
	return string(r.respBody)
}

// ToJSON convert json response body to struct or map
func (r *Req) ToJSON(v interface{}) error {
	return json.Unmarshal(r.respBody, v)
}

// ToXML convert xml response body to struct or map
func (r *Req) ToXML(v interface{}) error {
	return xml.Unmarshal(r.respBody, v)
}

// ToFile download the response body to file
func (r *Req) ToFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, r.resp.Body)
	if err != nil {
		return err
	}
	return nil
}

var regNewline = regexp.MustCompile(`\n|\r`)

func (r *Req) Format(s fmt.State, verb rune) {
	if r == nil || r.req == nil {
		return
	}
	req := r.req
	if s.Flag('+') { // include header and format pretty.
		fmt.Fprint(s, r.dump())
	} else if s.Flag('-') { // keep all informations in one line.
		fmt.Fprint(s, req.Method, " ", req.URL.String())
		if ShowCost {
			fmt.Fprint(s, " ", r.cost.String())
		}
		if len(r.reqBody) > 0 {
			str := regNewline.ReplaceAllString(string(r.reqBody), " ")
			fmt.Fprint(s, " ", str)
		}
		if str := string(r.respBody); str != "" {
			str = regNewline.ReplaceAllString(str, " ")
			fmt.Fprint(s, " ", str)
		}
	} else { // auto
		fmt.Fprint(s, req.Method, " ", req.URL.String())
		if ShowCost {
			fmt.Fprint(s, " ", r.cost.String())
		}
		respBody := r.respBody
		if (len(r.reqBody) > 0 && regNewline.Match(r.reqBody)) || (len(respBody) > 0 && regNewline.Match(respBody)) { // pretty format
			if len(r.reqBody) > 0 {
				fmt.Fprint(s, "\n", string(r.reqBody))
			}
			if len(respBody) > 0 {
				fmt.Fprint(s, "\n", string(respBody))
			}
		} else {
			if len(r.reqBody) > 0 {
				fmt.Fprint(s, " ", string(r.reqBody))
			}
			if len(respBody) > 0 {
				fmt.Fprint(s, " ", string(respBody))
			}
		}
	}

}

// Get execute a http GET request
func Get(url string, v ...interface{}) (*Req, error) {
	return Do("GET", url, v...)
}

// Post execute a http POST request
func Post(url string, v ...interface{}) (*Req, error) {
	return Do("POST", url, v...)
}

// Put execute a http PUT request
func Put(url string, v ...interface{}) (*Req, error) {
	return Do("PUT", url, v...)
}

// Patch execute a http PATCH request
func Patch(url string, v ...interface{}) (*Req, error) {
	return Do("PATCH", url, v...)
}

// Delete execute a http DELETE request
func Delete(url string, v ...interface{}) (*Req, error) {
	return Do("DELETE", url, v...)
}

// Head execute a http HEAD request
func Head(url string, v ...interface{}) (*Req, error) {
	return Do("HEAD", url, v...)
}

// Options execute a http OPTIONS request
func Options(url string, v ...interface{}) (*Req, error) {
	return Do("OPTIONS", url, v...)
}
