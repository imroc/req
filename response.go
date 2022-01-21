package req

import (
	"net/http"
	"strings"
)

type ResponseOptions struct {
	// DisableAutoDecode, if true, prevents auto detect response
	// body's charset and decode it to utf-8
	DisableAutoDecode bool

	// AutoDecodeContentType specifies an optional function for determine
	// whether the response body should been auto decode to utf-8.
	// Only valid when DisableAutoDecode is true.
	AutoDecodeContentType func(contentType string) bool

	Discard bool
}

type ResponseOption func(o *ResponseOptions)

func DiscardBody() ResponseOption {
	return func(o *ResponseOptions) {
		o.Discard = true
	}
}

// DisableAutoDecode disable the response body auto-decode to improve performance.
func DisableAutoDecode() ResponseOption {
	return func(o *ResponseOptions) {
		o.DisableAutoDecode = true
	}
}

// AutoDecodeContentTypeFunc customize the function to determine whether response
// body should auto decode with specified content type.
func AutoDecodeContentTypeFunc(fn func(contentType string) bool) ResponseOption {
	return func(o *ResponseOptions) {
		o.AutoDecodeContentType = fn
	}
}

// AutoDecodeContentType specifies that the response body should been auto-decoded
// when content type contains keywords that here given.
func AutoDecodeContentType(contentTypes ...string) ResponseOption {
	return func(o *ResponseOptions) {
		o.AutoDecodeContentType = func(contentType string) bool {
			for _, t := range contentTypes {
				if strings.Contains(contentType, t) {
					return true
				}
			}
			return false
		}
	}
}

type Response struct {
	*http.Response
	request *Request
}

func (r *Response) Body() Body {
	return Body{r.Response.Body, r.Response}
}
