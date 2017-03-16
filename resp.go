package req

import (
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

func (r *Response) Receive() (err error) {
	if r.Body != nil {
		return
	}
	defer r.Raw.Body.Close()
	r.Body, err = ioutil.ReadAll(r.Raw.Body)
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
	fmt.Fprint(s, string(r.Body))
}
