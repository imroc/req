package tests

import (
	"context"
	"crypto/tls"
	"github.com/imroc/req/v3/internal/dump"
	"github.com/imroc/req/v3/internal/transport"
	reqtls "github.com/imroc/req/v3/pkg/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Transport struct {
	ProxyValue                  func(*http.Request) (*url.URL, error)
	TLSClientConfigValue        *tls.Config
	DisableCompressionValue     bool
	DisableKeepAlivesValue      bool
	TLSHandshakeTimeoutValue    time.Duration
	ResponseHeaderTimeoutValue  time.Duration
	ExpectContinueTimeoutValue  time.Duration
	IdleConnTimeoutValue        time.Duration
	DumpValue                   *dump.Dumper
	ReadBufferSizeValue         int
	WriteBufferSizeValue        int
	MaxIdleConnsValue           int
	MaxIdleConnsPerHostValue    int
	MaxConnsPerHostValue        int
	MaxResponseHeaderBytesValue int64
	TLSNextProtoValue           map[string]func(authority string, c reqtls.Conn) http.RoundTripper
	GetProxyConnectHeaderValue  func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error)
	ProxyConnectHeaderValue     http.Header
	DebugfValue                 func(format string, v ...interface{})
}

func (t Transport) Proxy() func(*http.Request) (*url.URL, error) {
	return t.ProxyValue
}

func (t Transport) Clone() transport.Interface {
	return nil
}

func (t Transport) Debugf() func(format string, v ...interface{}) {
	return t.DebugfValue
}

func (t Transport) SetDebugf(f func(format string, v ...interface{})) {
	t.DebugfValue = f
}

func (t Transport) DisableCompression() bool {
	return t.DisableCompressionValue
}

func (t Transport) TLSClientConfig() *tls.Config {
	return t.TLSClientConfigValue
}

func (t Transport) SetTLSClientConfig(c *tls.Config) {
	t.TLSClientConfigValue = c
}

func (t Transport) TLSHandshakeTimeout() time.Duration {
	return t.TLSHandshakeTimeoutValue
}

func (t Transport) DialContext() func(ctx context.Context, network string, addr string) (net.Conn, error) {
	return nil
}

func (t Transport) DialTLSContext() func(ctx context.Context, network string, addr string) (net.Conn, error) {
	return nil
}

func (t Transport) RegisterProtocol(scheme string, rt http.RoundTripper) {
}

func (t Transport) DisableKeepAlives() bool {
	return t.DisableKeepAlivesValue
}

func (t Transport) Dump() *dump.Dumper {
	return t.DumpValue

}

func (t Transport) MaxIdleConns() int {
	return t.MaxIdleConnsValue
}

func (t Transport) MaxIdleConnsPerHost() int {
	return t.MaxIdleConnsPerHostValue
}

func (t Transport) MaxConnsPerHost() int {
	return t.MaxConnsPerHostValue
}

func (t Transport) IdleConnTimeout() time.Duration {
	return t.IdleConnTimeoutValue
}

func (t Transport) ResponseHeaderTimeout() time.Duration {
	return t.ResponseHeaderTimeoutValue
}

func (t Transport) ExpectContinueTimeout() time.Duration {
	return t.ExpectContinueTimeoutValue
}

func (t Transport) ProxyConnectHeader() http.Header {
	return t.ProxyConnectHeaderValue
}

func (t Transport) GetProxyConnectHeader() func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error) {
	return t.GetProxyConnectHeaderValue
}

func (t Transport) MaxResponseHeaderBytes() int64 {
	return t.MaxResponseHeaderBytesValue
}

func (t Transport) WriteBufferSize() int {
	return t.WriteBufferSizeValue
}

func (t Transport) ReadBufferSize() int {
	return t.ReadBufferSizeValue
}

func (t Transport) TLSNextProto() map[string]func(authority string, c reqtls.Conn) http.RoundTripper {
	return t.TLSNextProtoValue
}

func (t Transport) SetTLSNextProto(m map[string]func(authority string, c reqtls.Conn) http.RoundTripper) {
	t.TLSNextProtoValue = m
}
