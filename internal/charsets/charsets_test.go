// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package charsets

import (
	"github.com/imroc/req/v3/internal/tests"
	"os"
	"runtime"
	"testing"
)

var sniffTestCases = []struct {
	filename, want string
}{
	{"UTF-16LE-BOM.html", "utf-16le"},
	{"UTF-16BE-BOM.html", "utf-16be"},
	{"meta-content-attribute.html", "iso-8859-15"},
	{"meta-charset-attribute.html", "iso-8859-15"},
	{"HTTP-vs-UTF-8-BOM.html", "utf-8"},
	{"UTF-8-BOM-vs-meta-content.html", "utf-8"},
	{"UTF-8-BOM-vs-meta-charset.html", "utf-8"},
}

func TestSniff(t *testing.T) {
	switch runtime.GOOS {
	case "nacl": // platforms that don't permit direct file system access
		t.Skipf("not supported on %q", runtime.GOOS)
	}

	for _, tc := range sniffTestCases {
		content, err := os.ReadFile(tests.GetTestFilePath(tc.filename))
		if err != nil {
			t.Errorf("%s: error reading file: %v", tc.filename, err)
			continue
		}

		_, name := FindEncoding(content)
		if name != tc.want {
			t.Errorf("%s: got %q, want %q", tc.filename, name, tc.want)
			continue
		}
	}
}

var metaTestCases = []struct {
	meta, want string
}{
	{"", ""},
	{"text/html", ""},
	{"text/html; charset utf-8", ""},
	{"text/html; charset=latin-2", "latin-2"},
	{"text/html; charset; charset = utf-8", "utf-8"},
	{`charset="big5"`, "big5"},
	{"charset='shift_jis'", "shift_jis"},
}

func TestFromMeta(t *testing.T) {
	for _, tc := range metaTestCases {
		got := fromMetaElement(tc.meta)
		if got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.meta, got, tc.want)
		}
	}
}
