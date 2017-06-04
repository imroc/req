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
}

func (r *Resp) getRespBody() []byte {
	if r.respBody == nil {
		respBody, _ := ioutil.ReadAll(r.resp.Body)
		if respBody == nil {
			respBody = make([]byte, 0)
		}
		r.respBody = respBody
	}
	return r.respBody
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
	return r.getRespBody()
}

// String returns response body as string
func (r *Resp) String() string {
	return string(r.getRespBody())
}

// ToJSON convert json response body to struct or map
func (r *Resp) ToJSON(v interface{}) error {
	return json.Unmarshal(r.getRespBody(), v)
}

// ToXML convert xml response body to struct or map
func (r *Resp) ToXML(v interface{}) error {
	return xml.Unmarshal(r.getRespBody(), v)
}

// ToFile download the response body to file
func (r *Resp) ToFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	if r.respBody != nil {
		_, err = file.Write(r.respBody)
		if err != nil {
			return err
		}
	} else {
		_, err = io.Copy(file, r.resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

var regNewline = regexp.MustCompile(`\n|\r`)

func (r *Resp) Format(s fmt.State, verb rune) {
	if r == nil || r.req == nil {
		return
	}
	req := r.req
	if s.Flag('+') { // include header and format pretty.
		fmt.Fprint(s, r.dump())
	} else if s.Flag('-') { // keep all informations in one line.
		fmt.Fprint(s, req.Method, " ", req.URL.String())
		if r.r.flag&Lcost != 0 {
			fmt.Fprint(s, " ", r.cost.String())
		}
		if r.r.flag&LreqBody != 0 {
			if len(r.reqBody) > 0 {
				str := regNewline.ReplaceAllString(string(r.reqBody), " ")
				fmt.Fprint(s, " ", str)
			} else {
				fmt.Fprint(s, " ******")
			}
		}
		if r.r.flag&LrespBody != 0 {
			if respBody := r.String(); len(respBody) > 0 {
				str := regNewline.ReplaceAllString(respBody, " ")
				fmt.Fprint(s, " ", str)
			} else {
				fmt.Fprint(s, " ******")
			}
		}
	} else { // auto
		fmt.Fprint(s, req.Method, " ", req.URL.String())
		if r.r.flag&Lcost != 0 {
			fmt.Fprint(s, " ", r.cost.String())
		}
		var pretty bool
		if r.r.flag&LreqBody != 0 && len(r.reqBody) > 0 && regNewline.Match(r.reqBody) {
			pretty = true
		}
		if r.r.flag&LrespBody != 0 && len(r.getRespBody()) > 0 && regNewline.Match(r.getRespBody()) {
			pretty = true
		}
		if pretty {
			if len(r.reqBody) > 0 {
				fmt.Fprint(s, "\n", string(r.reqBody))
			}
			if len(r.respBody) > 0 {
				fmt.Fprint(s, "\n", string(r.respBody))
			}
		} else {
			if len(r.reqBody) > 0 {
				fmt.Fprint(s, " ", string(r.reqBody))
			}
			if len(r.respBody) > 0 {
				fmt.Fprint(s, " ", string(r.respBody))
			}
		}
	}

}
