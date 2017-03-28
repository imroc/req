package req

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	resp *http.Response
	body []byte
}

func NewResponse(resp *http.Response) (r *Response) {
	r = &Response{
		resp: resp,
	}
	return
}

var ErrNilResponse = errors.New("nil response")

func (r *Response) Response() *http.Response {
	if r == nil {
		return nil
	}
	return r.resp
}

func (r *Response) Receive() (err error) {
	if r == nil {
		err = ErrNilResponse
		return
	}
	if r.body != nil {
		return
	}
	defer r.resp.Body.Close()
	r.body, err = ioutil.ReadAll(r.resp.Body)
	return
}

func (r *Response) ReceiveBytes() (body []byte, err error) {
	if r == nil {
		err = ErrNilResponse
		return
	}
	if r.body == nil {
		if err = r.Receive(); err != nil {
			return
		}
	}
	body = r.body
	return
}

func (r *Response) Bytes() (body []byte) {
	body, _ = r.ReceiveBytes()
	return
}

func (r *Response) ReceiveString() (s string, err error) {
	if r == nil {
		err = ErrNilResponse
		return
	}
	if r.body == nil {
		if err = r.Receive(); err != nil {
			return
		}
	}
	s = string(r.body)
	return
}

func (r *Response) String() (s string) {
	s, _ = r.ReceiveString()
	return
}

func (r *Response) ToJson(v interface{}) (err error) {
	if r == nil {
		err = ErrNilResponse
		return
	}
	if r.body == nil {
		if err = r.Receive(); err != nil {
			return
		}
	}
	err = json.Unmarshal(r.body, v)
	return
}

func (r *Response) ToXml(v interface{}) (err error) {
	if r == nil {
		err = ErrNilResponse
		return
	}
	if r.body == nil {
		if err = r.Receive(); err != nil {
			return
		}
	}
	err = xml.Unmarshal(r.body, v)
	return
}

func (r *Response) Format(s fmt.State, verb rune) {
	if r == nil {
		return
	}
	if r.resp == nil {
		return
	}
	str := r.String()
	if str == "" {
		return
	}
	resp := r.resp
	if s.Flag('+') {
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
		fmt.Fprint(s, "\n\n", str)
		return
	} else if s.Flag('-') {
		str = regBlank.ReplaceAllString(str, "")
		fmt.Fprint(s, str)
		return
	}
	fmt.Fprint(s, str)
	return
}
