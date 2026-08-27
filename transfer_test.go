package req

import (
	"bufio"
	"net/http"
	"strings"
	"testing"
)

func TestReadTransferClosesResponseWithTransferEncodingAndContentLength(t *testing.T) {
	res := &http.Response{
		StatusCode: 200,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Transfer-Encoding": {"chunked"},
			"Content-Length":    {"3"},
		},
	}

	err := readTransfer(res, bufio.NewReader(strings.NewReader("3\r\nfoo\r\n0\r\n\r\n")))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Close {
		t.Fatal("response with both Transfer-Encoding and Content-Length did not set Close")
	}
}
