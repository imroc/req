package req

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Response is the http response.
type Response struct {
	*http.Response
	Request    *Request
	body       []byte
	receivedAt time.Time
}

// IsSuccess method returns true if HTTP status `code >= 200 and <= 299` otherwise false.
func (r *Response) IsSuccess() bool {
	return r.StatusCode > 199 && r.StatusCode < 300
}

// IsError method returns true if HTTP status `code >= 400` otherwise false.
func (r *Response) IsError() bool {
	return r.StatusCode > 399
}

// GetContentType return the `Content-Type` header value.
func (r *Response) GetContentType() string {
	return r.Header.Get(hdrContentTypeKey)
}

// Result returns the response value as an object if it has one
func (r *Response) Result() interface{} {
	return r.Request.Result
}

// Error returns the error object if it has one.
func (r *Response) Error() interface{} {
	return r.Request.Error
}

// TraceInfo returns the TraceInfo from Request.
func (r *Response) TraceInfo() TraceInfo {
	return r.Request.TraceInfo()
}

func (r *Response) TotalTime() time.Duration {
	if r.Request.trace != nil {
		return r.Request.TraceInfo().TotalTime
	}
	return r.receivedAt.Sub(r.Request.StartTime)
}

func (r *Response) ReceivedAt() time.Time {
	return r.receivedAt
}

func (r *Response) setReceivedAt() {
	r.receivedAt = time.Now()
	if r.Request.trace != nil {
		r.Request.trace.endTime = r.receivedAt
	}
}
func (r *Response) UnmarshalJson(v interface{}) error {
	b, err := r.ToBytes()
	if err != nil {
		return err
	}
	return r.Request.client.jsonUnmarshal(b, v)
}

func (r *Response) UnmarshalXml(v interface{}) error {
	b, err := r.ToBytes()
	if err != nil {
		return err
	}
	return r.Request.client.xmlUnmarshal(b, v)
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
