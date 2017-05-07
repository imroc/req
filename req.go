package req

import (
	"bytes"
	"crypto/tls"
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

// custom http client
var Client *http.Client

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
}

func getClient() *http.Client {
	if Client != nil {
		return Client
	}
	return defaultClient
}

func getTransport() *http.Transport {
	trans, _ := getClient().Transport.(*http.Transport)
	return trans
}

// EnableInsecureTLS
func EnableInsecureTLS(enable bool) {
	trans := getTransport()
	if trans == nil {
		return
	}
	if trans.TLSClientConfig == nil {
		trans.TLSClientConfig = &tls.Config{}
	}
	trans.TLSClientConfig.InsecureSkipVerify = enable
}

// enable or disable cookie manager
func EnableCookie(enable bool) {
	if enable {
		jar, _ := cookiejar.New(nil)
		getClient().Jar = jar
	} else {
		getClient().Jar = nil
	}
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
		if b.Content == nil {
			if b.Data == nil {
				return
			}
			b.Content = ioutil.NopCloser(bytes.NewReader(b.Data))
		}
		req.Body = b.Content
		if b.ContentType != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", b.ContentType)
		}
		if b.Data != nil {
			r.reqBody = b.Data
			req.ContentLength = int64(len(b.Data))
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
			if rc, ok := t.(io.ReadCloser); ok {
				req.Body = rc
			} else {
				req.Body = ioutil.NopCloser(t)
			}
		case *body:
			handleBody(t)
		case Param:
			param = append(param, t)
		case string:
			handleBody(&body{Content: ioutil.NopCloser(strings.NewReader(t)), Data: []byte(t)})
		case []byte:
			handleBody(&body{Content: ioutil.NopCloser(bytes.NewReader(t)), Data: t})
		case *http.Client:
			r.client = t
		case FileUpload:
			file = append(file, t)
		case []FileUpload:
			if file == nil {
				file = make([]FileUpload, len(t))
			}
			copy(file, t)
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
			for _, p := range param {
				for key, value := range p {
					bodyWriter.WriteField(key, value)
				}
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
			body := &body{
				ContentType: "application/x-www-form-urlencoded",
				Data:        []byte(paramStr),
				Content:     ioutil.NopCloser(strings.NewReader(paramStr)),
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
		if Client != nil {
			r.client = Client
		} else {
			r.client = defaultClient
		}
	}

	resp, errDo := r.client.Do(req)
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
	return
}

// Body represents request's body
type body struct {
	ContentType string
	Content     io.ReadCloser
	Data        []byte
}

// BodyXML get request's body as xml
func BodyXML(v interface{}) interface{} {
	b := new(body)
	switch t := v.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		b.Content = ioutil.NopCloser(bf)
		b.Data = []byte(t)
	case []byte:
		bf := bytes.NewBuffer(t)
		b.Content = ioutil.NopCloser(bf)
		b.Data = t
	default:
		bs, err := xml.Marshal(v)
		if err != nil {
			return err
		}
		bf := bytes.NewBuffer(bs)
		b.Content = ioutil.NopCloser(bf)
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
		bf := bytes.NewBufferString(t)
		b.Content = ioutil.NopCloser(bf)
		b.Data = []byte(t)
	case []byte:
		bf := bytes.NewBuffer(t)
		b.Content = ioutil.NopCloser(bf)
		b.Data = t
	default:
		bs, err := json.Marshal(v)
		if err != nil {
			return err
		}
		bf := bytes.NewBuffer(bs)
		b.Content = ioutil.NopCloser(bf)
		b.Data = bs
	}
	b.ContentType = "text/json"
	return b
}

// File upload file of the specified filename.
func File(filename string) interface{} {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	return FileUpload{
		File:     file,
		FileName: filepath.Base(filename),
	}
}

// FileGlob upload files matching the name pattern such as
// /usr/*/bin/go* (assuming the Separator is '/')
func FileGlob(pattern string) interface{} {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return errors.New("req: No files have been matched")
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

var regBlank = regexp.MustCompile(`\s+`)

func (r *Req) Format(s fmt.State, verb rune) {
	if r == nil || r.req == nil {
		return
	}
	req := r.req
	if s.Flag('+') { // include header and format pretty.
		fmt.Fprint(s, req.Method, " ", req.URL.String(), " ", req.Proto)
		for name, values := range req.Header {
			for _, value := range values {
				fmt.Fprint(s, "\n", name, ":", value)
			}
		}
		if len(r.reqBody) > 0 {
			fmt.Fprint(s, "\n\n", string(r.reqBody))
		}
		if r.resp != nil {
			resp := r.resp
			fmt.Fprint(s, "\n\n")
			fmt.Fprint(s, resp.Proto, " ", resp.Status) // e.g. HTTP/1.1 200 OK
			//header
			if len(resp.Header) > 0 {
				for name, values := range resp.Header {
					for _, value := range values {
						fmt.Fprintf(s, "\n%s:%s", name, value)
					}
				}
			}
			//body
			fmt.Fprint(s, "\n\n", string(r.respBody))
		}
	} else if s.Flag('-') { // keep all informations in one line.
		fmt.Fprint(s, req.Method, " ", req.URL.String())
		if len(r.reqBody) > 0 {
			str := regBlank.ReplaceAllString(string(r.reqBody), "")
			fmt.Fprint(s, str)
		}
		if str := string(r.reqBody); str != "" {
			str = regBlank.ReplaceAllString(str, "")
			fmt.Fprint(s, " ", str)
		}
	} else { // auto
		fmt.Fprint(s, req.Method, " ", req.URL.String())
		respBody := r.respBody
		if (len(r.reqBody) > 0 && bytes.IndexByte(r.reqBody, '\n') != -1) || (len(respBody) > 0 && bytes.IndexByte(respBody, '\n') != -1) { // pretty format
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
