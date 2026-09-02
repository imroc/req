package req

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
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

func TestResponseWithBareLFInTrailerClosesConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	var connections, requests atomic.Int32
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			connections.Add(1)
			go func() {
				defer conn.Close()
				r := bufio.NewReader(conn)
				for {
					req, err := http.ReadRequest(r)
					if err != nil {
						return
					}
					req.Body.Close()
					if requests.Add(1) == 1 {
						fmt.Fprint(conn, "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\nTrailer: X-One, X-Two\r\n\r\n0\r\nX-One: a\nX-Two: b\r\n\r\n")
					} else {
						fmt.Fprint(conn, "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
					}
				}
			}()
		}
	}()

	tr := NewTransport()
	defer tr.CloseIdleConnections()
	client := &http.Client{Transport: tr}
	resp, err := client.Get("http://" + ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err == nil || err.Error() != "http: invalid trailer" {
		t.Fatalf("got body read error %v, want http: invalid trailer", err)
	}

	resp, err = client.Get("http://" + ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Fatal(err)
	}
	if got := connections.Load(); got != 2 {
		t.Fatalf("got %d connections, want 2; invalid trailer connection was reused", got)
	}
}
