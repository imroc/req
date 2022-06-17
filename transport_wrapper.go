package req

import (
	"context"
	"crypto/tls"
	"github.com/imroc/req/v3/internal/dump"
	reqtls "github.com/imroc/req/v3/internal/tls"
	"github.com/imroc/req/v3/internal/transport"
	"net"
	"net/http"
	"net/url"
	"time"
)

type transportImpl struct {
	t *Transport
}

func (t transportImpl) Proxy() func(*http.Request) (*url.URL, error) {
	return t.t.Proxy
}

func (t transportImpl) Clone() transport.Interface {
	return transportImpl{t.t.Clone()}
}

func (t transportImpl) Debugf() func(format string, v ...interface{}) {
	return t.t.Debugf
}

func (t transportImpl) SetDebugf(f func(format string, v ...interface{})) {
	t.t.Debugf = f
}

func (t transportImpl) DisableCompression() bool {
	return t.t.DisableCompression
}

func (t transportImpl) TLSClientConfig() *tls.Config {
	return t.t.TLSClientConfig
}

func (t transportImpl) SetTLSClientConfig(c *tls.Config) {
	t.t.TLSClientConfig = c
}

func (t transportImpl) TLSHandshakeTimeout() time.Duration {
	return t.t.TLSHandshakeTimeout
}

func (t transportImpl) DialContext() func(ctx context.Context, network string, addr string) (net.Conn, error) {
	return t.t.DialContext
}

func (t transportImpl) DialTLSContext() func(ctx context.Context, network string, addr string) (net.Conn, error) {
	return t.t.DialTLSContext
}

func (t transportImpl) RegisterProtocol(scheme string, rt http.RoundTripper) {
	t.t.RegisterProtocol(scheme, rt)
}

func (t transportImpl) DisableKeepAlives() bool {
	return t.t.DisableKeepAlives
}

func (t transportImpl) Dump() *dump.Dumper {
	return t.t.dump
}

func (t transportImpl) MaxIdleConns() int {
	return t.t.MaxIdleConns
}

func (t transportImpl) MaxIdleConnsPerHost() int {
	return t.t.MaxIdleConnsPerHost
}

func (t transportImpl) MaxConnsPerHost() int {
	return t.t.MaxConnsPerHost
}

func (t transportImpl) IdleConnTimeout() time.Duration {
	return t.t.IdleConnTimeout
}

func (t transportImpl) ResponseHeaderTimeout() time.Duration {
	return t.t.ResponseHeaderTimeout
}

func (t transportImpl) ExpectContinueTimeout() time.Duration {
	return t.t.ExpectContinueTimeout
}

func (t transportImpl) ProxyConnectHeader() http.Header {
	return t.t.ProxyConnectHeader
}

func (t transportImpl) GetProxyConnectHeader() func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error) {
	return t.t.GetProxyConnectHeader
}

func (t transportImpl) MaxResponseHeaderBytes() int64 {
	return t.t.MaxResponseHeaderBytes
}

func (t transportImpl) WriteBufferSize() int {
	return t.t.WriteBufferSize
}

func (t transportImpl) ReadBufferSize() int {
	return t.t.ReadBufferSize
}

func (t transportImpl) TLSNextProto() map[string]func(authority string, c reqtls.Conn) http.RoundTripper {
	return t.t.TLSNextProto
}

func (t transportImpl) SetTLSNextProto(m map[string]func(authority string, c reqtls.Conn) http.RoundTripper) {
	t.t.TLSNextProto = m
}
