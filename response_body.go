package req

import (
	"io/ioutil"
	"strings"
)

func (r *Response) UnmarshalJson(v interface{}) error {
	b, err := r.ToBytes()
	if err != nil {
		return err
	}
	return r.Request.client.JSONUnmarshal(b, v)
}

func (r *Response) UnmarshalXml(v interface{}) error {
	b, err := r.ToBytes()
	if err != nil {
		return err
	}
	return r.Request.client.XMLUnmarshal(b, v)
}

func (r *Response) Unmarshal(v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "json") {
		return r.UnmarshalJson(v)
	} else if strings.Contains(contentType, "xml") {
		return r.UnmarshalXml(v)
	}
	return r.UnmarshalJson(v)
}

// Bytes return the response body as []bytes that hava already been read, could be
// nil if not read, the following cases are already read:
// 1. `Request.SetResult` or `Request.SetError` is called.
// 2. `Client.DisableAutoReadResponse(false)` is not called,
//     also `Request.SetOutput` and `Request.SetOutputFile` is not called.
func (r *Response) Bytes() []byte {
	return r.body
}

// String return the response body as string that hava already been read, could be
// nil if not read, the following cases are already read:
// 1. `Request.SetResult` or `Request.SetError` is called.
// 2. `Client.DisableAutoReadResponse(false)` is not called,
//     also `Request.SetOutput` and `Request.SetOutputFile` is not called.
func (r *Response) String() string {
	return string(r.body)
}

func (r *Response) ToString() (string, error) {
	b, err := r.ToBytes()
	return string(b), err
}

func (r *Response) ToBytes() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	r.setReceivedAt()
	if err != nil {
		return nil, err
	}
	r.body = body
	return body, nil
}
