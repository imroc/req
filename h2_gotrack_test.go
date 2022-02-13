// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package req

import (
	"fmt"
	"strings"
	"testing"
)

func TestGoroutineLock(t *testing.T) {
	oldDebug := http2DebugGoroutines
	http2DebugGoroutines = true
	defer func() { http2DebugGoroutines = oldDebug }()

	g := http2newGoroutineLock()
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
