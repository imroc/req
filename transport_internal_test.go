// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// White-box tests for transport.go (in package http instead of http_test).

package req

import (
	"context"
	"errors"
	"github.com/imroc/req/v3/internal/http2"
	"github.com/imroc/req/v3/internal/tests"
	"net"
	"net/http"
	"strings"
	"testing"
)

func withT(r *http.Request, t *testing.T) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), tLogKey{}, t.Logf))
}

// Issue 15446: incorrect wrapping of errors when server closes an idle connection.
func TestTransportPersistConnReadLoopEOF(t *testing.T) {
	ln := tests.NewLocalListener(t)
	defer ln.Close()

	connc := make(chan net.Conn, 1)
	go func() {
		defer close(connc)
		c, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		connc <- c
	}()

	tr := new(Transport)
	req, _ := http.NewRequest("GET", "http://"+ln.Addr().String(), nil)
	req = withT(req, t)
	treq := &transportRequest{Request: req}
	cm := connectMethod{targetScheme: "http", targetAddr: ln.Addr().String()}
	pc, err := tr.getConn(treq, cm)
	if err != nil {
		t.Fatal(err)
	}
	defer pc.close(errors.New("test over"))

	conn := <-connc
	if conn == nil {
		// Already called t.Error in the accept goroutine.
		return
	}
	conn.Close() // simulate the server hanging up on the client

	_, err = pc.roundTrip(treq)
	if !isNothingWrittenError(err) && !isTransportReadFromServerError(err) && err != errServerClosedIdle {
		t.Errorf("roundTrip = %#v, %v; want errServerClosedIdle, transportReadFromServerError, or nothingWrittenError", err, err)
	}

	<-pc.closech
	err = pc.closed
	if !isTransportReadFromServerError(err) && err != errServerClosedIdle {
		t.Errorf("pc.closed = %#v, %v; want errServerClosedIdle or transportReadFromServerError", err, err)
	}
}

func isNothingWrittenError(err error) bool {
	_, ok := err.(nothingWrittenError)
	return ok
}

func isTransportReadFromServerError(err error) bool {
	_, ok := err.(transportReadFromServerError)
	return ok
}

func dummyRequest(method string) *http.Request {
	req, err := http.NewRequest(method, "http://fake.tld/", nil)
	if err != nil {
		panic(err)
	}
	return req
}
func dummyRequestWithBody(method string) *http.Request {
	req, err := http.NewRequest(method, "http://fake.tld/", strings.NewReader("foo"))
	if err != nil {
		panic(err)
	}
	return req
}

func dummyRequestWithBodyNoGetBody(method string) *http.Request {
	req := dummyRequestWithBody(method)
	req.GetBody = nil
	return req
}

// issue22091Error acts like a golang.org/x/net/http2.ErrNoCachedConn.
type issue22091Error struct{}

func (issue22091Error) IsHTTP2NoCachedConnError() {}
func (issue22091Error) Error() string             { return "issue22091Error" }

func TestTransportShouldRetryRequest(t *testing.T) {
	tests := []struct {
		pc  *persistConn
		req *http.Request

		err  error
		want bool
	}{
		0: {
			pc:   &persistConn{reused: false},
			req:  dummyRequest("POST"),
			err:  nothingWrittenError{},
			want: false,
		},
		1: {
			pc:   &persistConn{reused: true},
			req:  dummyRequest("POST"),
			err:  nothingWrittenError{},
			want: true,
		},
		2: {
			pc:   &persistConn{reused: true},
			req:  dummyRequest("POST"),
			err:  http2.ErrNoCachedConn,
			want: true,
		},
		3: {
			pc:   nil,
			req:  nil,
			err:  issue22091Error{}, // like an external http2ErrNoCachedConn
			want: true,
		},
		4: {
			pc:   &persistConn{reused: true},
			req:  dummyRequest("POST"),
			err:  errMissingHost,
			want: false,
		},
		5: {
			pc:   &persistConn{reused: true},
			req:  dummyRequest("POST"),
			err:  transportReadFromServerError{},
			want: false,
		},
		6: {
			pc:   &persistConn{reused: true},
			req:  dummyRequest("GET"),
			err:  transportReadFromServerError{},
			want: true,
		},
		7: {
			pc:   &persistConn{reused: true},
			req:  dummyRequest("GET"),
			err:  errServerClosedIdle,
			want: true,
		},
		8: {
			pc:   &persistConn{reused: true},
			req:  dummyRequestWithBody("POST"),
			err:  nothingWrittenError{},
			want: true,
		},
		9: {
			pc:   &persistConn{reused: true},
			req:  dummyRequestWithBodyNoGetBody("POST"),
			err:  nothingWrittenError{},
			want: false,
		},
	}
	for i, tt := range tests {
		got := tt.pc.shouldRetryRequest(tt.req, tt.err)
		if got != tt.want {
			t.Errorf("%d. shouldRetryRequest = %v; want %v", i, got, tt.want)
		}
	}
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
