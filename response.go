package req

import (
	"net/http"
	"time"
)

// ResponseOptions determines that how should the response been processed.
type ResponseOptions struct {
	// DisableAutoDecode, if true, prevents auto detect response
	// body's charset and decode it to utf-8
	DisableAutoDecode bool

	// AutoDecodeContentType specifies an optional function for determine
	// whether the response body should been auto decode to utf-8.
	// Only valid when DisableAutoDecode is true.
	AutoDecodeContentType func(contentType string) bool
}

var textContentTypes = []string{"text", "json", "xml", "html", "java"}

var autoDecodeText = autoDecodeContentTypeFunc(textContentTypes...)

func autoDecodeContentTypeFunc(contentTypes ...string) func(contentType string) bool {
	m := make(map[string]bool)
	for _, ct := range contentTypes {
		m[ct] = true
	}
	return func(contentType string) bool {
		return m[contentType]
	}
}

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

func (r *Response) GetContentType() string {
	return r.Header.Get(hdrContentTypeKey)
}

// Result method returns the response value as an object if it has one
func (r *Response) Result() interface{} {
	return r.Request.Result
}

// Error method returns the error object if it has one
func (r *Response) Error() interface{} {
	return r.Request.Error
}

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
