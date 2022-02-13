// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package req

import (
	"net/http/httptrace"
	"net/textproto"
)

func http2traceHasWroteHeaderField(trace *httptrace.ClientTrace) bool {
	return trace != nil && trace.WroteHeaderField != nil
}

func http2traceWroteHeaderField(trace *httptrace.ClientTrace, k, v string) {
	if trace != nil && trace.WroteHeaderField != nil {
		trace.WroteHeaderField(k, []string{v})
	}
}

func http2traceGot1xxResponseFunc(trace *httptrace.ClientTrace) func(int, textproto.MIMEHeader) error {
	if trace != nil {
		return trace.Got1xxResponse
	}
	return nil
}
