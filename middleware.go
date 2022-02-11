package req

import (
	"bytes"
	"github.com/imroc/req/v3/internal/util"
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

func createMultipartHeader(file *FileUpload, contentType string) textproto.MIMEHeader {
	hdr := make(textproto.MIMEHeader)

	contentDispositionValue := "form-data"
	cd := new(ContentDisposition)
	if file.ParamName != "" {
		cd.Add("name", file.ParamName)
	}
	if file.FileName != "" {
		cd.Add("filename", file.FileName)
	}
	if file.ExtraContentDisposition != nil {
		for _, kv := range file.ExtraContentDisposition.kv {
			cd.Add(kv.Key, kv.Value)
		}
	}
	if c := cd.String(); c != "" {
		contentDispositionValue += c
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

func writeMultipartFormFile(w *multipart.Writer, file *FileUpload) error {
	defer closeq(file.File)
	// Auto detect actual multipart content type
	cbuf := make([]byte, 512)
	size, err := file.File.Read(cbuf)
	if err != nil && err != io.EOF {
		return err
	}

	pw, err := w.CreatePart(createMultipartHeader(file, http.DetectContentType(cbuf)))
	if err != nil {
		return err
	}

	if _, err = pw.Write(cbuf[:size]); err != nil {
		return err
	}

	_, err = io.Copy(pw, file.File)
	return err
}

func writeMultiPart(r *Request, w *multipart.Writer, pw *io.PipeWriter) {
	for k, vs := range r.FormData {
		for _, v := range vs {
			w.WriteField(k, v)
		}
	}
	for _, file := range r.uploadFiles {
		writeMultipartFormFile(w, file)
	}
	w.Close()  // close multipart to write tailer boundary
	pw.Close() // close pipe writer so that pipe reader could get EOF, and stop upload
}

func handleMultiPart(c *Client, r *Request) (err error) {
	pr, pw := io.Pipe()
	r.RawRequest.Body = pr
	w := multipart.NewWriter(pw)
	r.SetContentType(w.FormDataContentType())
	go writeMultiPart(r, w, pw)
	return
}

func handleFormData(r *Request) {
	r.SetContentType(formContentType)
	r.SetBodyBytes([]byte(r.FormData.Encode()))
}

func handleMarshalBody(c *Client, r *Request) error {
	ct := ""
	if r.Headers != nil {
		ct = r.Headers.Get(hdrContentTypeKey)
	}
	if ct == "" {
		ct = c.Headers.Get(hdrContentTypeKey)
	}
	if ct != "" {
		if util.IsXMLType(ct) {
			body, err := c.xmlMarshal(r.marshalBody)
			if err != nil {
				return err
			}
			r.SetBodyBytes(body)
		} else {
			body, err := c.jsonMarshal(r.marshalBody)
			if err != nil {
				return err
			}
			r.SetBodyBytes(body)
		}
		return nil
	}
	body, err := c.jsonMarshal(r.marshalBody)
	if err != nil {
		return err
	}
	r.SetBodyJsonBytes(body)
	return nil
}

func parseRequestBody(c *Client, r *Request) (err error) {
	if c.isPayloadForbid(r.RawRequest.Method) {
		r.RawRequest.Body = nil
		return
	}
	// handle multipart
	if r.isMultiPart && (r.RawRequest.Method != http.MethodPatch) {
		return handleMultiPart(c, r)
	}

	// handle form data
	if len(c.FormData) > 0 {
		r.SetFormDataFromValues(c.FormData)
	}
	if len(r.FormData) > 0 {
		handleFormData(r)
		return
	}

	// handle marshal body
	if r.marshalBody != nil {
		handleMarshalBody(c, r)
	}

	if r.getHeader(hdrContentTypeKey) == "" && r.body != nil {
		ct := http.DetectContentType(r.body)
		if ct != "application/octet-stream" {
			r.SetContentType(ct)
		}
	}
	return
}

func unmarshalBody(c *Client, r *Response, v interface{}) (err error) {
	body, err := r.ToBytes() // in case req.SetResult or req.SetError with cient.DisalbeAutoReadResponse(true)
	if err != nil {
		return
	}
	ct := r.GetContentType()
	if util.IsJSONType(ct) {
		return c.jsonUnmarshal(body, v)
	} else if util.IsXMLType(ct) {
		return c.xmlUnmarshal(body, v)
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

	var output io.Writer
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
		closeq(output)
	}()
	_, err = io.Copy(output, body)
	r.setReceivedAt()
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

	// If Request.URL is relative path then added c.BaseURL into
	// the request URL otherwise Request.URL will be used as-is
	if !reqURL.IsAbs() {
		r.URL = reqURL.String()
		if len(r.URL) > 0 && r.URL[0] != '/' {
			r.URL = "/" + r.URL
		}

		reqURL, err = url.Parse(c.BaseURL + r.URL)
		if err != nil {
			return err
		}
	}

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
