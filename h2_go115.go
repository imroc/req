// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.15
// +build go1.15

package req

import (
	"context"
	"crypto/tls"
)

// dialTLSWithContext uses tls.Dialer, added in Go 1.15, to open a TLS
// connection.
func (t *http2Transport) dialTLSWithContext(ctx context.Context, network, addr string, cfg *tls.Config) (TLSConn, error) {
	dialer := &tls.Dialer{
		Config: cfg,
	}
	cn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	tlsCn := cn.(TLSConn) // DialContext comment promises this will always succeed
	return tlsCn, nil
}
