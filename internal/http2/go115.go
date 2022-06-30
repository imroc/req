// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.15
// +build go1.15

package http2

import (
	"context"
	"crypto/tls"
	reqtls "github.com/imroc/req/v3/pkg/tls"
)

// dialTLSWithContext uses tls.Dialer, added in Go 1.15, to open a TLS
// connection.
func (t *Transport) dialTLSWithContext(ctx context.Context, network, addr string, cfg *tls.Config) (reqtls.Conn, error) {
	dialer := &tls.Dialer{
		Config: cfg,
	}
	cn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	tlsCn := cn.(reqtls.Conn) // DialContext comment promises this will always succeed
	return tlsCn, nil
}
