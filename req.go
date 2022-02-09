package req

import (
	"fmt"
	"io"
)

const (
	hdrUserAgentKey   = "User-Agent"
	hdrUserAgentValue = "req/v3 (https://github.com/imroc/req)"
	hdrContentTypeKey = "Content-Type"
	plainTextType     = "text/plain; charset=utf-8"
	jsonContentType   = "application/json; charset=utf-8"
	xmlContentType    = "text/xml; charset=utf-8"
	formContentType   = "application/x-www-form-urlencoded"
)

type kv struct {
	Key   string
	Value string
}

// ContentDisposition represents parameters in `Content-Disposition`
// MIME header of multipart request.
type ContentDisposition struct {
	kv []kv
}

func (c *ContentDisposition) Add(key, value string) *ContentDisposition {
	c.kv = append(c.kv, kv{Key: key, Value: value})
	return c
}

func (c *ContentDisposition) String() string {
	if c == nil {
		return ""
	}
	s := ""
	for _, kv := range c.kv {
		s += fmt.Sprintf("; %s=%q", kv.Key, kv.Value)
	}
	return s
}

// FileUpload represents a "form-data" multipart
type FileUpload struct {
	// "name" parameter in `Content-Disposition`
	ParamName string
	// "filename" parameter in `Content-Disposition`
	FileName string
	// The file to be uploaded.
	File io.Reader

	// According to the HTTP specification, this should be nil,
	// but some servers may not follow the specification and
	// requires `Content-Disposition` parameters more than just
	// "name" and "filename".
	ExtraContentDisposition *ContentDisposition // Usually
}
