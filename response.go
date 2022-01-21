package req

import (
	"io"
	"io/ioutil"
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

	// AutoDiscard, if true, read all response body and discard automatically,
	// useful when test
	AutoDiscard bool
}

type ResponseOption func(o *ResponseOptions)

func DiscardResponseBody() ResponseOption {
	return func(o *ResponseOptions) {
		o.AutoDiscard = true
	}
}

// DisableAutoDecode disable the response body auto-decode to improve performance.
func DisableAutoDecode() ResponseOption {
	return func(o *ResponseOptions) {
		o.DisableAutoDecode = true
	}
}

var textContentTypes = []string{"text", "json", "xml", "html", "java"}

func AutoDecodeTextContent() ResponseOption {
	return AutoDecodeContentType(textContentTypes...)
}

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

// AutoDecodeContentType specifies that the response body should been auto-decoded
// when content type contains keywords that here given.
func AutoDecodeContentType(contentTypes ...string) ResponseOption {
	return func(o *ResponseOptions) {
		o.AutoDecodeContentType = autoDecodeContentTypeFunc(contentTypes...)
	}
}

type Response struct {
	*http.Response
	request *Request
}

func (r *Response) Body() Body {
	return Body{r.Response.Body, r.Response}
}

func (r *Response) Discard() error {
	_, err := io.Copy(ioutil.Discard, r.Response.Body)
	return err
}
