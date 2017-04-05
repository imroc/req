package req

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Response wraps *http.Response, makes it much easier to handle response.
type Response struct {
	resp *http.Response
	body []byte
}

// WrapResponse create a Response which wraps *http.Response
func WrapResponse(resp *http.Response) (r *Response) {
	r = &Response{
		resp: resp,
	}
	return
}

var ErrNilResponse = errors.New("nil response")

// GetResponse return the raw *http.Response inside the Response.
func (r *Response) GetResponse() *http.Response {
	if r == nil {
		return nil
	}
	return r.resp
}

// Receive accept the body of http response if haven't been accepted yet,
// err is not nil if error happens during read the response body.
func (r *Response) Receive() (err error) {
	if r == nil || r.resp == nil {
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

// ReceiveBytes accept the body of http response if haven't been accepted yet,
// return the response body as []byte, err is not nil if error happens during read the response body.
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

// Bytes accept the body of http response if haven't been accepted yet,
// return the response body as []byte, body is nil if error happens.
func (r *Response) Bytes() (body []byte) {
	body, _ = r.ReceiveBytes()
	return
}

// ReceiveString accept the body of http response if haven't been accepted yet,
// return the response body as string, err is not nil if error happens during read the response body.
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

// String accept the body of http response if haven't been accepted yet,
// return the response body as string, if error happens, s is "".
func (r *Response) String() (s string) {
	s, _ = r.ReceiveString()
	return
}

// ToXML accept the body of http response if haven't been acceptted yet, v is the address
// of the struct you want to unmarshal, err is not nil when error happens.
func (r *Response) ToJSON(v interface{}) (err error) {
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

// ToXML accept the body of http response if haven't been acceptted yet, v is the address
// of the struct you want to unmarshal, err is not nil when error happens.
func (r *Response) ToXML(v interface{}) (err error) {
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
