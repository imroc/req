package requtil

import (
	"bytes"
	"net/http"
)

// ConvertHeaderToString converts http header to a string.
func ConvertHeaderToString(h http.Header) string {
	buf := new(bytes.Buffer)
	h.Write(buf)
	return buf.String()
}
