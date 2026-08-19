package req

import (
	"net/http"
	"sort"
	"strings"
)

// GenerateCurlCommand generates a curl command that is equivalent to the
// request, which is handy for debugging: the returned command can be pasted
// into a terminal to reproduce the request outside of the program.
//
// If the request has already been sent, the actual request that was fired is
// used. Otherwise the request-building middlewares are run against a copy of
// the request to resolve the final URL, headers, cookies and body, so calling
// GenerateCurlCommand before sending does not alter the request. The returned
// command reflects what would be sent, including query parameters, path
// parameters, form data, a marshaled JSON/XML body and basic/bearer auth and
// cookie headers.
//
// Some transport-level details that are only decided when the request is
// actually sent are not represented, such as the automatically added
// Accept-Encoding header, HTTP/2 or HTTP/3 negotiation, digest auth (which
// needs a challenge from the server) and streaming multipart file uploads
// whose body is not buffered in memory.
func (r *Request) GenerateCurlCommand() string {
	if r == nil || r.client == nil {
		return ""
	}

	// The request was already sent, reuse the raw request that was fired.
	if r.RawRequest != nil {
		return buildCurlCommand(r.RawRequest, r.Body)
	}

	// The request has not been sent yet, run the request-building middlewares
	// on a copy so the original request is left untouched (avoids, for example,
	// cookies being appended twice or the body being marshaled twice when the
	// same request is later sent).
	rc := *r
	rc.Headers = r.Headers.Clone()
	if rc.Headers == nil {
		rc.Headers = make(http.Header)
	}
	rc.Cookies = append([]*http.Cookie(nil), r.Cookies...)

	for _, f := range rc.client.udBeforeRequest {
		if err := f(rc.client, &rc); err != nil {
			return ""
		}
	}
	for _, f := range rc.client.beforeRequest {
		if err := f(rc.client, &rc); err != nil {
			return ""
		}
	}
	if rc.URL == nil {
		return ""
	}

	req := &http.Request{
		Method: rc.Method,
		URL:    rc.URL,
		Header: rc.Headers.Clone(),
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	for _, cookie := range rc.Cookies {
		req.AddCookie(cookie)
	}
	return buildCurlCommand(req, rc.Body)
}

// buildCurlCommand renders req (and its in-memory body) as a curl command.
func buildCurlCommand(req *http.Request, body []byte) string {
	if req == nil || req.URL == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("curl")

	method := req.Method
	if method == "" {
		method = http.MethodGet
	}
	// curl defaults to GET, and to POST when a body is supplied, so only emit
	// -X when the method cannot be inferred from the rest of the command.
	if method != http.MethodGet || len(body) > 0 {
		b.WriteString(" -X ")
		b.WriteString(method)
	}

	b.WriteByte(' ')
	b.WriteString(shellEscape(req.URL.String()))

	// Emit headers in a stable order so the output is deterministic.
	keys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range req.Header[k] {
			b.WriteString(" -H ")
			b.WriteString(shellEscape(k + ": " + v))
		}
	}

	// A Host header set on an already-sent request lives on Request.Host rather
	// than in the header map, add it back so the command stays faithful.
	if req.Host != "" && req.Host != req.URL.Host && req.Header.Get("Host") == "" {
		b.WriteString(" -H ")
		b.WriteString(shellEscape("Host: " + req.Host))
	}

	if len(body) > 0 {
		b.WriteString(" -d ")
		b.WriteString(shellEscape(string(body)))
	}

	return b.String()
}

// shellEscape wraps s in single quotes so it is safe to paste into a POSIX
// shell, escaping any embedded single quotes.
func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
