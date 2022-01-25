package req

import (
	"net/http"
	"strings"
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
	return func(contentType string) bool {
		for _, t := range contentTypes {
			if strings.Contains(contentType, t) {
				return true
			}
		}
		return false
	}
}

// Response is the http response.
type Response struct {
	*http.Response
	Request *Request
	body    []byte
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
