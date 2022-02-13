// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package req

import (
	"flag"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func init() {
	http2inTests = true
	http2DebugGoroutines = true
	flag.BoolVar(&http2VerboseLogs, "verboseh2", http2VerboseLogs, "Verbose HTTP/2 debug logging")
}

func TestSettingString(t *testing.T) {
	tests := []struct {
		s    http2Setting
		want string
	}{
		{http2Setting{http2SettingMaxFrameSize, 123}, "[MAX_FRAME_SIZE = 123]"},
		{http2Setting{1<<16 - 1, 123}, "[UNKNOWN_SETTING_65535 = 123]"},
	}
	for i, tt := range tests {
		got := fmt.Sprint(tt.s)
		if got != tt.want {
			t.Errorf("%d. for %#v, string = %q; want %q", i, tt.s, got, tt.want)
		}
	}
}

func cleanDate(res *http.Response) {
	if d := res.Header["Date"]; len(d) == 1 {
		d[0] = "XXX"
	}
}

func TestSorterPoolAllocs(t *testing.T) {
	ss := []string{"a", "b", "c"}
	h := http.Header{
		"a": nil,
		"b": nil,
		"c": nil,
	}
	sorter := new(http2sorter)

	if allocs := testing.AllocsPerRun(100, func() {
		sorter.SortStrings(ss)
	}); allocs >= 1 {
		t.Logf("SortStrings allocs = %v; want <1", allocs)
	}

	if allocs := testing.AllocsPerRun(5, func() {
		if len(sorter.Keys(h)) != 3 {
			t.Fatal("wrong result")
		}
	}); allocs > 0 {
		t.Logf("Keys allocs = %v; want <1", allocs)
	}
}

// waitCondition reports whether fn eventually returned true,
// checking immediately and then every checkEvery amount,
// until waitFor has elapsed, at which point it returns false.
func waitCondition(waitFor, checkEvery time.Duration, fn func() bool) bool {
	deadline := time.Now().Add(waitFor)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(checkEvery)
	}
	return false
}

// waitErrCondition is like waitCondition but with errors instead of bools.
func waitErrCondition(waitFor, checkEvery time.Duration, fn func() error) error {
	deadline := time.Now().Add(waitFor)
	var err error
	for time.Now().Before(deadline) {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(checkEvery)
	}
	return err
}
