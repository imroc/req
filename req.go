package req

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// default *Req
var std = New()

// flags to decide which part can be outputed
const (
	LreqHead  = 1 << iota // output request head (request line and request header)
	LreqBody              // output request body
	LrespHead             // output response head (response line and response header)
	LrespBody             // output response body
	Lcost                 // output time costed by the request
	LstdFlags = LreqHead | LreqBody | LrespHead | LrespBody
)

// Req is a convenient client for initiating requests
type Req struct {
	client      *http.Client
	jsonEncOpts *jsonEncOpts
	xmlEncOpts  *xmlEncOpts
	flag        int
}

// New create a new *Req
func New() *Req {
	return &Req{flag: LstdFlags}
}

var regTextContentType = regexp.MustCompile("text|xml|json|javascript|charset|java")

// Do execute a http request with sepecify method and url,
// and it can also have some optional params, depending on your needs.
func (r *Req) Do(method, rawurl string, v ...interface{}) (resp *Resp, err error) {
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
	resp = &Resp{req: req, r: r}
	handleBody := func(data []byte, contentType string) {
		resp.reqBody = data
		req.Body = resp.getReqBody()
		req.ContentLength = int64(len(resp.reqBody))
		if contentType != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	var formParam []Param
	var queryParam []QueryParam
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
			var rc io.ReadCloser
			if trc, ok := t.(io.ReadCloser); ok {
				rc = trc
			} else {
				rc = ioutil.NopCloser(t)
			}
			req.Body = bodyWrapper{
				ReadCloser: rc,
				limit:      102400,
			}
			bs, err := ioutil.ReadAll(t)
			if err != nil {
				return nil, err
			}
			handleBody(bs, "")
			if rc, ok := t.(io.ReadCloser); ok {
				rc.Close()
			}
		case *bodyJson:
			var data []byte
			if r.jsonEncOpts != nil {
				opts := r.jsonEncOpts
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetIndent(opts.indentPrefix, opts.indentValue)
				enc.SetEscapeHTML(opts.escapeHTML)
				err = enc.Encode(t.v)
				if err != nil {
					return nil, err
				}
				data = buf.Bytes()
			} else {
				data, err = json.Marshal(t.v)
				if err != nil {
					return nil, err
				}
			}
			handleBody(data, "application/json; charset=UTF-8")
		case *bodyXml:
			var data []byte
			if r.xmlEncOpts != nil {
				opts := r.xmlEncOpts
				var buf bytes.Buffer
				enc := xml.NewEncoder(&buf)
				enc.Indent(opts.prefix, opts.indent)
				err = enc.Encode(t.v)
				if err != nil {
					return nil, err
				}
				data = buf.Bytes()
			} else {
				data, err = xml.Marshal(t.v)
				if err != nil {
					return nil, err
				}
			}
			handleBody(data, "application/xml; charset=UTF-8")
		case Param:
			if method == "GET" {
				queryParam = append(queryParam, QueryParam(t))
			} else {
				formParam = append(formParam, t)
			}
		case QueryParam:
			queryParam = append(queryParam, t)
		case string:
			handleBody([]byte(t), "")
		case []byte:
			handleBody(t, "")
		case *http.Client:
			resp.client = t
		case FileUpload:
			file = append(file, t)
		case []FileUpload:
			if file == nil {
				file = make([]FileUpload, 0)
			}
			file = append(file, t...)
		case *http.Cookie:
			req.AddCookie(t)
		case Host:
			req.Host = string(t)
		case error:
			err = t
			return
		}
	}

	if len(file) > 0 && (req.Method == "POST" || req.Method == "PUT") {
		r.upload(resp, file, formParam)
	} else if len(formParam) > 0 {
		if req.Body != nil {
			return nil, errors.New("req: can not set both body and params")
		}
		params := make(url.Values)
		for _, p := range formParam {
			for key, value := range p {
				params.Add(key, value)
			}
		}
		paramStr := params.Encode()
		handleBody([]byte(paramStr), "application/x-www-form-urlencoded; charset=UTF-8")
	}

	if len(queryParam) > 0 {
		params := make(url.Values)
		for _, p := range queryParam {
			for key, value := range p {
				params.Add(key, value)
			}
		}
		paramStr := params.Encode()
		if strings.IndexByte(rawurl, '?') == -1 {
			rawurl = rawurl + "?" + paramStr
		} else {
			rawurl = rawurl + "&" + paramStr
		}
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req.URL = u

	if resp.client == nil {
		resp.client = r.Client()
	}

	now := time.Now()
	response, err := resp.client.Do(req)
	resp.cost = time.Since(now)
	if err != nil {
		return
	}

	resp.resp = response
	ct := response.Header.Get("Content-Type")
	if ct == "" || regTextContentType.MatchString(ct) { // text
		defer response.Body.Close()
		var reader io.Reader
		if _, ok := resp.client.Transport.(*http.Transport); ok && response.Header.Get("Content-Encoding") == "gzip" && req.Header.Get("Accept-Encoding") != "" {
			reader, err = gzip.NewReader(response.Body)
			if err != nil {
				return nil, err
			}
		} else {
			reader = response.Body
		}
		respBody, err := ioutil.ReadAll(reader)
		if err != nil {
			return resp, err
		}
		resp.respBody = respBody
	}
	if Debug {
		fmt.Println(resp.dump())
	}
	return
}

type bodyWrapper struct {
	io.ReadCloser
	buf   bytes.Buffer
	limit int
}

func (b bodyWrapper) Read(p []byte) (n int, err error) {
	n, err = b.ReadCloser.Read(p)
	if left := b.limit - b.buf.Len(); left > 0 && n > 0 {
		if n <= left {
			b.buf.Write(p[:n])
		} else {
			b.buf.Write(p[:left])
		}
	}
	return
}

type dummyMultipart struct {
	buf bytes.Buffer
	w   *multipart.Writer
}

func (d *dummyMultipart) WriteField(fieldname, value string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, fieldname))
	p, err := d.w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = p.Write([]byte(value))
	return err
}

func (d *dummyMultipart) WriteFile(fieldname, filename string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			fieldname, filename))
	h.Set("Content-Type", "application/octet-stream")
	p, err := d.w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = p.Write([]byte("******"))
	return err
}

func newDummyMultipart() *dummyMultipart {
	d := new(dummyMultipart)
	d.w = multipart.NewWriter(&d.buf)
	return d
}

func (r *Req) upload(resp *Resp, file []FileUpload, param []Param) {
	pr, pw := io.Pipe()
	bodyWriter := multipart.NewWriter(pw)
	d := newDummyMultipart()
	go func() {
		for _, p := range param {
			for key, value := range p {
				bodyWriter.WriteField(key, value)
				d.WriteField(key, value)
			}
		}
		i := 0
		for _, f := range file {
			if f.FieldName == "" {
				i++
				f.FieldName = "file" + strconv.Itoa(i)
			}
			fileWriter, err := bodyWriter.CreateFormFile(f.FieldName, f.FileName)
			if err != nil {
				return
			}
			//iocopy
			_, err = io.Copy(fileWriter, f.File)
			if err != nil {
				return
			}
			f.File.Close()
			d.WriteFile(f.FieldName, f.FileName)
		}
		bodyWriter.Close()
		pw.Close()
		resp.reqBody = d.buf.Bytes()
	}()
	resp.req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	resp.req.Body = ioutil.NopCloser(pr)
}

// Get execute a http GET request
func (r *Req) Get(url string, v ...interface{}) (*Resp, error) {
	return r.Do("GET", url, v...)
}

// Post execute a http POST request
func (r *Req) Post(url string, v ...interface{}) (*Resp, error) {
	return r.Do("POST", url, v...)
}

// Put execute a http PUT request
func (r *Req) Put(url string, v ...interface{}) (*Resp, error) {
	return r.Do("PUT", url, v...)
}

// Patch execute a http PATCH request
func (r *Req) Patch(url string, v ...interface{}) (*Resp, error) {
	return r.Do("PATCH", url, v...)
}

// Delete execute a http DELETE request
func (r *Req) Delete(url string, v ...interface{}) (*Resp, error) {
	return r.Do("DELETE", url, v...)
}

// Head execute a http HEAD request
func (r *Req) Head(url string, v ...interface{}) (*Resp, error) {
	return r.Do("HEAD", url, v...)
}

// Options execute a http OPTIONS request
func (r *Req) Options(url string, v ...interface{}) (*Resp, error) {
	return r.Do("OPTIONS", url, v...)
}

// Get execute a http GET request
func Get(url string, v ...interface{}) (*Resp, error) {
	return std.Get(url, v...)
}

// Post execute a http POST request
func Post(url string, v ...interface{}) (*Resp, error) {
	return std.Post(url, v...)
}

// Put execute a http PUT request
func Put(url string, v ...interface{}) (*Resp, error) {
	return std.Put(url, v...)
}

// Head execute a http HEAD request
func Head(url string, v ...interface{}) (*Resp, error) {
	return std.Head(url, v...)
}

// Options execute a http OPTIONS request
func Options(url string, v ...interface{}) (*Resp, error) {
	return std.Options(url, v...)
}

// Delete execute a http DELETE request
func Delete(url string, v ...interface{}) (*Resp, error) {
	return std.Delete(url, v...)
}

// Patch execute a http PATCH request
func Patch(url string, v ...interface{}) (*Resp, error) {
	return std.Patch(url, v...)
}

// Do execute request.
func Do(method, url string, v ...interface{}) (*Resp, error) {
	return std.Do(method, url, v...)
}
