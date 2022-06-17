// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http2

import (
	"fmt"
	"github.com/imroc/req/v3/internal/tests"
	"strings"
	"testing"
)

func TestGoroutineLock(t *testing.T) {
	oldDebug := DebugGoroutines
	DebugGoroutines = true
	defer func() { DebugGoroutines = oldDebug }()

	g := newGoroutineLock()
	g.check()

	sawPanic := make(chan interface{})
	go func() {
		defer func() { sawPanic <- recover() }()
		g.check() // should panic
	}()
	e := <-sawPanic
	if e == nil {
		t.Fatal("did not see panic from check in other goroutine")
	}
	if !strings.Contains(fmt.Sprint(e), "wrong goroutine") {
		t.Errorf("expected on see panic about running on the wrong goroutine; got %v", e)
	}
}

func TestParseUintBytes(t *testing.T) {
	s := []byte{}
	_, err := parseUintBytes(s, 0, 0)
	tests.AssertErrorContains(t, err, "invalid syntax")

	s = []byte("0x")
	_, err = parseUintBytes(s, 0, 0)
	tests.AssertErrorContains(t, err, "invalid syntax")

	s = []byte("0x01")
	_, err = parseUintBytes(s, 0, 0)
	tests.AssertNoError(t, err)

	s = []byte("0xa1")
	_, err = parseUintBytes(s, 0, 0)
	tests.AssertNoError(t, err)

	s = []byte("0xA1")
	_, err = parseUintBytes(s, 0, 0)
	tests.AssertNoError(t, err)
}
