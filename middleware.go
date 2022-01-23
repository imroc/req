package req

import (
	"bytes"
	"github.com/imroc/req/v2/internal/util"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type (
	// RequestMiddleware type is for request middleware, called before a request is sent
	RequestMiddleware func(*Client, *Request) error

	// ResponseMiddleware type is for response middleware, called after a response has been received
	ResponseMiddleware func(*Client, *Response) error
)

// unmarshalc content into object from JSON or XML
func unmarshalc(c *Client, ct string, b []byte, d interface{}) (err error) {
	if util.IsJSONType(ct) {
		err = c.JSONUnmarshal(b, d)
	} else if util.IsXMLType(ct) {
		err = c.XMLUnmarshal(b, d)
	}
	return
}

func parseResponseBody(c *Client, r *Response) (err error) {
	if r.StatusCode == http.StatusNoContent {
		return
	}
	body, err := r.Bytes() // in case req.SetResult with cient.DisalbeAutoReadResponse(true)
	if err != nil {
		return
	}
	// Handles only JSON or XML content type
	ct := util.FirstNonEmpty(r.GetContentType())
	if r.IsSuccess() && r.Request.Result != nil {
		r.Request.Error = nil
		if util.IsJSONType(ct) {
			return c.JSONUnmarshal(body, r.Request.Result)
		} else if util.IsXMLType(ct) {
			return c.XMLUnmarshal(body, r.Request.Result)
		}
	}
	if r.IsError() && r.Request.Error != nil {
		r.Request.Result = nil
		if util.IsJSONType(ct) {
			return c.JSONUnmarshal(body, r.Request.Error)
		} else if util.IsXMLType(ct) {
			return c.XMLUnmarshal(body, r.Request.Error)
		}
	}
	return
}

func handleDownload(c *Client, r *Response) error {
	if !r.Request.isSaveResponse {
		return nil
	}

	var body io.ReadCloser
	if r.body != nil { // already read
		body = ioutil.NopCloser(bytes.NewReader(r.body))
	} else {
		body = r.Body
	}

	defer func() {
		body.Close()
		r.Request.output.Close()
	}()
	_, err := io.Copy(r.Request.output, body)
	return err
}

func parseRequestHeader(c *Client, r *Request) error {
	if c.Headers == nil {
		return nil
	}
	hdr := make(http.Header)
	for k := range c.Headers {
		hdr[k] = append(hdr[k], c.Headers[k]...)
	}

	for k := range r.Headers {
		hdr.Del(k)
		hdr[k] = append(hdr[k], r.Headers[k]...)
	}

	r.Headers = hdr
	return nil
}

func parseRequestURL(c *Client, r *Request) error {
	if len(r.PathParams) > 0 {
		for p, v := range r.PathParams {
			r.URL = strings.Replace(r.URL, "{"+p+"}", url.PathEscape(v), -1)
		}
	}
	if len(c.PathParams) > 0 {
		for p, v := range c.PathParams {
			r.URL = strings.Replace(r.URL, "{"+p+"}", url.PathEscape(v), -1)
		}
	}

	// Parsing request URL
	reqURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	// If Request.URL is relative path then added c.HostURL into
	// the request URL otherwise Request.URL will be used as-is
	if !reqURL.IsAbs() {
		r.URL = reqURL.String()
		if len(r.URL) > 0 && r.URL[0] != '/' {
			r.URL = "/" + r.URL
		}

		reqURL, err = url.Parse(c.HostURL + r.URL)
		if err != nil {
			return err
		}
	}

	// GH #407 && #318
	if reqURL.Scheme == "" && len(c.scheme) > 0 {
		reqURL.Scheme = c.scheme
	}

	// Adding Query Param
	query := make(url.Values)
	for k, v := range c.QueryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	for k, v := range r.QueryParams {
		// remove query param from client level by key
		// since overrides happens for that key in the request
		query.Del(k)

		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Preserve query string order partially.
	// Since not feasible in `SetQuery*` resty methods, because
	// standard package `url.Encode(...)` sorts the query params
	// alphabetically
	if len(query) > 0 {
		if util.IsStringEmpty(reqURL.RawQuery) {
			reqURL.RawQuery = query.Encode()
		} else {
			reqURL.RawQuery = reqURL.RawQuery + "&" + query.Encode()
		}
	}

	r.URL = reqURL.String()

	return nil
}
