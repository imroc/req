package req

import (
	"bytes"
	"fmt"
	"github.com/imroc/req/v2/internal/util"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type (
	// RequestMiddleware type is for request middleware, called before a request is sent
	RequestMiddleware func(*Client, *Request) error

	// ResponseMiddleware type is for response middleware, called after a response has been received
	ResponseMiddleware func(*Client, *Response) error
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func createMultipartHeader(param, fileName, contentType string) textproto.MIMEHeader {
	hdr := make(textproto.MIMEHeader)

	var contentDispositionValue string
	if util.IsStringEmpty(fileName) {
		contentDispositionValue = fmt.Sprintf(`form-data; name="%s"`, param)
	} else {
		contentDispositionValue = fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			param, escapeQuotes(fileName))
	}
	hdr.Set("Content-Disposition", contentDispositionValue)

	if !util.IsStringEmpty(contentType) {
		hdr.Set(hdrContentTypeKey, contentType)
	}
	return hdr
}

func closeq(v interface{}) {
	if c, ok := v.(io.Closer); ok {
		c.Close()
	}
}

type multipartBody struct {
	*io.PipeReader
	closed bool
}

func (r *multipartBody) Read(p []byte) (n int, err error) {
	if r.closed {
		return 0, io.EOF
	}
	n, err = r.PipeReader.Read(p)
	if err != nil {
		r.closed = true
		err = nil
	}
	return
}

func writeMultipartFormFile(w *multipart.Writer, fieldName, fileName string, r io.Reader) error {
	defer closeq(r)
	// Auto detect actual multipart content type
	cbuf := make([]byte, 512)
	size, err := r.Read(cbuf)
	if err != nil && err != io.EOF {
		return err
	}

	pw, err := w.CreatePart(createMultipartHeader(fieldName, fileName, http.DetectContentType(cbuf)))
	if err != nil {
		return err
	}

	if _, err = pw.Write(cbuf[:size]); err != nil {
		return err
	}

	_, err = io.Copy(pw, r)
	return err
}

func writeMultiPart(c *Client, r *Request, w *multipart.Writer, pw *io.PipeWriter) {
	for k, vs := range r.FormData {
		for _, v := range vs {
			w.WriteField(k, v)
		}
	}
	for _, file := range r.uploadFiles {
		writeMultipartFormFile(w, file.ParamName, file.FilePath, file.Reader)
	}
	w.Close()  // close multipart to write tailer boundary
	pw.Close() // close pipe writer so that pipe reader could get EOF, and stop upload
}

func handleMultiPart(c *Client, r *Request) (err error) {
	pr, pw := io.Pipe()
	r.RawRequest.Body = pr
	w := multipart.NewWriter(pw)
	r.RawRequest.Header.Set(hdrContentTypeKey, w.FormDataContentType())
	go writeMultiPart(c, r, w, pw)
	return
}

func handleFormData(r *Request) {
	r.RawRequest.Body = ioutil.NopCloser(strings.NewReader(r.FormData.Encode()))
}

func parseRequestBody(c *Client, r *Request) (err error) {
	if r.isMultiPart {
		return handleMultiPart(c, r)
	}
	if len(r.FormData) > 0 {
		handleFormData(r)
		return
	}
	return
}

func unmarshalBody(c *Client, r *Response, v interface{}) (err error) {
	body, err := r.ToBytes() // in case req.SetResult or req.SetError with cient.DisalbeAutoReadResponse(true)
	if err != nil {
		return
	}
	ct := util.FirstNonEmpty(r.GetContentType())
	if util.IsJSONType(ct) {
		return c.JSONUnmarshal(body, v)
	} else if util.IsXMLType(ct) {
		return c.XMLUnmarshal(body, v)
	}
	return
}

func parseResponseBody(c *Client, r *Response) (err error) {
	if r.StatusCode == http.StatusNoContent {
		return
	}
	// Handles only JSON or XML content type
	if r.Request.Result != nil && r.IsSuccess() {
		unmarshalBody(c, r, r.Request.Result)
	}
	if r.Request.Error != nil && r.IsError() {
		unmarshalBody(c, r, r.Request.Error)
	}
	return
}

func handleDownload(c *Client, r *Response) (err error) {
	if !r.Request.isSaveResponse {
		return nil
	}
	var body io.ReadCloser

	if r.body != nil { // already read
		body = ioutil.NopCloser(bytes.NewReader(r.body))
	} else {
		body = r.Body
	}

	var output io.WriteCloser
	if r.Request.outputFile != "" {
		file := r.Request.outputFile
		if c.outputDirectory != "" && !filepath.IsAbs(file) {
			file = c.outputDirectory + string(filepath.Separator) + file
		}

		file = filepath.Clean(file)

		if err = util.CreateDirectory(filepath.Dir(file)); err != nil {
			return err
		}
		output, err = os.Create(file)
		if err != nil {
			return
		}
	} else {
		output = r.Request.output // must not nil
	}

	defer func() {
		body.Close()
		output.Close()
	}()
	_, err = io.Copy(output, body)
	return
}

func parseRequestHeader(c *Client, r *Request) error {
	if c.Headers == nil {
		return nil
	}
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	for k := range c.Headers {
		if r.Headers.Get(k) == "" {
			r.Headers.Add(k, c.Headers.Get(k))
		}
	}
	return nil
}

func parseRequestCookie(c *Client, r *Request) error {
	if len(c.Cookies) == 0 {
		return nil
	}
	for _, ck := range c.Cookies {
		r.Cookies = append(r.Cookies, ck)
	}
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
