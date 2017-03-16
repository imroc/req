package req

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Raw  *http.Response
	Body []byte
}

func WrapResponse(raw *http.Response) (r Response) {
	r = Response{
		Raw: raw,
	}
	return
}

var ErrNilResponse = errors.New("nil response")

func (r *Response) Receive() (err error) {
	if r.Body != nil {
		return
	}
	if r.Raw == nil {
		err = ErrNilResponse
		return
	}
	defer r.Raw.Body.Close()
	r.Body, err = ioutil.ReadAll(r.Raw.Body)
	return
}

func (r *Response) ToJson(v interface{}) (err error) {
	err = r.Receive()
	if err != nil {
		return
	}
	err = json.Unmarshal(r.Body, v)
	return
}

func (r *Response) ToXml(v interface{}) (err error) {
	err = r.Receive()
	if err != nil {
		return
	}
	err = xml.Unmarshal(r.Body, v)
	return
}

func (r Response) Format(s fmt.State, verb rune) {
	if r.Raw == nil {
		return
	}
	resp := r.Raw
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
		fmt.Fprint(s, "\n\n")
		//body
		r.Receive() // ensure body has received
		if r.Body == nil {
			return
		}
		fmt.Fprint(s, string(r.Body))
		return
	}
	if bytes.IndexByte(r.Body, '\n') != -1 {
		fmt.Fprint(s, "\n")
	}
	fmt.Fprint(s, string(r.Body))
}
