package req

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

// Resp represents a request with it's response
type Resp struct {
	r      *Req
	req    *http.Request
	resp   *http.Response
	client *http.Client
	*multipartHelper
	reqBody  []byte
	respBody []byte
	cost     time.Duration
	err      error // delayed error
}

// Cost returns time spent by the request
func (r *Resp) Cost() time.Duration {
	return r.cost
}

// Request returns *http.Request
func (r *Resp) Request() *http.Request {
	return r.req
}

// Response returns *http.Response
func (r *Resp) Response() *http.Response {
	return r.resp
}

// Bytes returns response body as []byte
func (r *Resp) Bytes() []byte {
	data, _ := r.ToBytes()
	return data
}

// ToBytes returns response body as []byte,
// return error if error happend when reading
// the response body
func (r *Resp) ToBytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.respBody != nil {
		return r.respBody, nil
	}
	respBody, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		r.err = err
		return nil, err
	}
	r.respBody = respBody
	return r.respBody, nil
}

// String returns response body as string
func (r *Resp) String() string {
	data, _ := r.ToBytes()
	return string(data)
}

// ToString returns response body as string,
// return error if error happend when reading
// the response body
func (r *Resp) ToString() (string, error) {
	data, err := r.ToBytes()
	return string(data), err
}

// ToJSON convert json response body to struct or map
func (r *Resp) ToJSON(v interface{}) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML convert xml response body to struct or map
func (r *Resp) ToXML(v interface{}) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToFile download the response body to file with optional download callback
func (r *Resp) ToFile(name string, callback ...DownloadCallback) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	if r.respBody != nil {
		_, err = file.Write(r.respBody)
		return err
	}

	if len(callback) > 0 && r.resp.ContentLength > 0 {
		return r.download(file, r.resp.ContentLength, callback...)
	}

	_, err = io.Copy(file, r.resp.Body)
	return err
}

func (r *Resp) download(file *os.File, totalLength int64, callback ...DownloadCallback) error {
	p := make([]byte, 1024)
	b := r.resp.Body
	var currentLength int64
	for {
		l, err := b.Read(p)
		if l > 0 {
			_, _err := file.Write(p[:l])
			if _err != nil {
				return _err
			}
			currentLength += int64(l)
			for _, cb := range callback {
				cb.OnProgress(currentLength, totalLength)
			}
		}
		if err != nil {
			if err == io.EOF {
				for _, cb := range callback {
					cb.OnComplete(file)
				}
				return nil
			}
			return err
		}
	}
}

var regNewline = regexp.MustCompile(`\n|\r`)

func (r *Resp) autoFormat(s fmt.State) {
	req := r.req
	fmt.Fprint(s, req.Method, " ", req.URL.String())
	if r.r.flag&Lcost != 0 {
		fmt.Fprint(s, " ", r.cost.String())
	}

	// test if it is should be outputed pretty
	var pretty bool
	var parts []string
	addPart := func(part string) {
		if part == "" {
			return
		}
		parts = append(parts, part)
		if !pretty && regNewline.MatchString(part) {
			pretty = true
		}
	}
	if r.r.flag&LreqBody != 0 { // request body
		addPart(string(r.reqBody))
	}
	if r.r.flag&LrespBody != 0 { // response body
		addPart(r.String())
	}

	for _, part := range parts {
		if pretty {
			fmt.Fprint(s, "\n")
		}
		fmt.Fprint(s, " ", part)
	}
}

func (r *Resp) miniFormat(s fmt.State) {
	req := r.req
	fmt.Fprint(s, req.Method, " ", req.URL.String())
	if r.r.flag&Lcost != 0 {
		fmt.Fprint(s, " ", r.cost.String())
	}
	if r.r.flag&LreqBody != 0 && len(r.reqBody) > 0 { // request body
		str := regNewline.ReplaceAllString(string(r.reqBody), " ")
		fmt.Fprint(s, " ", str)
	}
	if r.r.flag&LrespBody != 0 && r.String() != "" { // response body
		str := regNewline.ReplaceAllString(r.String(), " ")
		fmt.Fprint(s, " ", str)
	}
}

func (r *Resp) Format(s fmt.State, verb rune) {
	if r == nil || r.req == nil {
		return
	}
	if s.Flag('+') { // include header and format pretty.
		fmt.Fprint(s, r.dump())
	} else if s.Flag('-') { // keep all informations in one line.
		r.miniFormat(s)
	} else { // auto
		r.autoFormat(s)
	}
}

// DownloadCallback is the callback when downloading
type DownloadCallback interface {
	OnComplete(file *os.File)
	OnProgress(currentLength int64, totalLength int64)
}
