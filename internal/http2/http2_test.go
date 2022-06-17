// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http2

import (
	"flag"
	"fmt"
	"github.com/imroc/req/v3/internal/tests"
	"net/http"
	"testing"
	"time"
)

func init() {
	inTests = true
	DebugGoroutines = true
	flag.BoolVar(&VerboseLogs, "verboseh2", VerboseLogs, "Verbose HTTP/2 debug logging")
}

func TestSettingString(t *testing.T) {
	tests := []struct {
		s    Setting
		want string
	}{
		{Setting{SettingMaxFrameSize, 123}, "[MAX_FRAME_SIZE = 123]"},
		{Setting{1<<16 - 1, 123}, "[UNKNOWN_SETTING_65535 = 123]"},
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
	sorter := new(sorter)

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

func TestSettingValid(t *testing.T) {
	cases := []struct {
		id  SettingID
		val uint32
	}{
		{
			id:  SettingEnablePush,
			val: 2,
		},
		{
			id:  SettingInitialWindowSize,
			val: 1 << 31,
		},
		{
			id:  SettingMaxFrameSize,
			val: 0,
		},
	}
	for _, c := range cases {
		s := &Setting{ID: c.id, Val: c.val}
		tests.AssertEqual(t, true, s.Valid() != nil)
	}
	s := &Setting{ID: SettingMaxHeaderListSize}
	tests.AssertEqual(t, true, s.Valid() == nil)
}

func TestBodyAllowedForStatus(t *testing.T) {
	tests.AssertEqual(t, false, bodyAllowedForStatus(101))
	tests.AssertEqual(t, false, bodyAllowedForStatus(204))
	tests.AssertEqual(t, false, bodyAllowedForStatus(304))
	tests.AssertEqual(t, true, bodyAllowedForStatus(900))
}

func TestHttpError(t *testing.T) {
	e := &httpError{msg: "test"}
	tests.AssertEqual(t, "test", e.Error())
	tests.AssertEqual(t, true, e.Temporary())
}
