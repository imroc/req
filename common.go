package req

import "strings"

const (
	CONTENT_TYPE_APPLICATION_JSON_UTF8 = "application/json; charset=UTF-8"
	CONTENT_TYPE_APPLICATION_XML_UTF8  = "application/xml; charset=UTF-8"
	CONTENT_TYPE_TEXT_XML_UTF8         = "text/xml; charset=UTF-8"
	CONTENT_TYPE_TEXT_HTML_UTF8        = "text/html; charset=UTF-8"
	CONTENT_TYPE_TEXT_PLAIN_UTF8       = "text/plain; charset=UTF-8"
)


// cutString is a string util function which is copied
// from go1.18 strings package.
// cutString slices s around the first instance of sep,
// returning the text before and after sep.
// The found result reports whether sep appears in s.
// If sep does not appear in s, cut returns s, "", false.
func cutString(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}
