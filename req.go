package req

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	hdrUserAgentKey      = "User-Agent"
	hdrUserAgentValue    = "req/v3 (https://github.com/imroc/req)"
	hdrLocationKey       = "Location"
	hdrContentTypeKey    = "Content-Type"
	plainTextContentType = "text/plain; charset=utf-8"
	jsonContentType      = "application/json; charset=utf-8"
	xmlContentType       = "text/xml; charset=utf-8"
	formContentType      = "application/x-www-form-urlencoded"
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
	ExtraContentDisposition *ContentDisposition
}

func cloneCookies(cookies []*http.Cookie) []*http.Cookie {
	if len(cookies) == 0 {
		return nil
	}
	c := make([]*http.Cookie, len(cookies))
	copy(c, cookies)
	return c
}

func cloneHeaders(hdrs http.Header) http.Header {
	if hdrs == nil {
		return nil
	}
	h := make(http.Header)
	for k, vs := range hdrs {
		for _, v := range vs {
			h.Add(k, v)
		}
	}
	return h
}

// TODO: change to generics function when generics are commonly used.
func cloneRequestMiddleware(m []RequestMiddleware) []RequestMiddleware {
	if len(m) == 0 {
		return nil
	}
	mm := make([]RequestMiddleware, len(m))
	copy(mm, m)
	return mm
}

func cloneResponseMiddleware(m []ResponseMiddleware) []ResponseMiddleware {
	if len(m) == 0 {
		return nil
	}
	mm := make([]ResponseMiddleware, len(m))
	copy(mm, m)
	return mm
}

func cloneUrlValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}
	vv := make(url.Values)
	for key, values := range v {
		for _, value := range values {
			vv.Add(key, value)
		}
	}
	return vv
}

func cloneMap(h map[string]string) map[string]string {
	if h == nil {
		return nil
	}
	m := make(map[string]string)
	for k, v := range h {
		m[k] = v
	}
	return m
}
