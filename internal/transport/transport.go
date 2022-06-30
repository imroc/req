package transport

import (
	"context"
	"crypto/tls"
	"github.com/imroc/req/v3/internal/dump"
	reqtls "github.com/imroc/req/v3/internal/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Interface interface {
	Proxy() func(*http.Request) (*url.URL, error)
	Debugf() func(format string, v ...interface{})
	SetDebugf(func(format string, v ...interface{}))
	DisableCompression() bool
	TLSClientConfig() *tls.Config
	SetTLSClientConfig(c *tls.Config)
	TLSHandshakeTimeout() time.Duration
	DialContext() func(ctx context.Context, network, addr string) (net.Conn, error)
	DialTLSContext() func(ctx context.Context, network, addr string) (net.Conn, error)
	DisableKeepAlives() bool
	Dump() *dump.Dumper

	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns() int

	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// defaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost() int

	// MaxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	//
	// Zero means no limit.
	MaxConnsPerHost() int

	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	IdleConnTimeout() time.Duration

	// ResponseHeaderTimeout, if non-zero, specifies the amount of
	// time to wait for a server's response headers after fully
	// writing the request (including its body, if any). This
	// time does not include the time to read the response body.
	ResponseHeaderTimeout() time.Duration

	// ExpectContinueTimeout, if non-zero, specifies the amount of
	// time to wait for a server's first response headers after fully
	// writing the request headers if the request has an
	// "Expect: 100-continue" header. Zero means no timeout and
	// causes the body to be sent immediately, without
	// waiting for the server to approve.
	// This time does not include the time to send the request header.
	ExpectContinueTimeout() time.Duration

	// ProxyConnectHeader optionally specifies headers to send to
	// proxies during CONNECT requests.
	// To set the header dynamically, see GetProxyConnectHeader.
	ProxyConnectHeader() http.Header

	// GetProxyConnectHeader optionally specifies a func to return
	// headers to send to proxyURL during a CONNECT request to the
	// ip:port target.
	// If it returns an error, the Transport's RoundTrip fails with
	// that error. It can return (nil, nil) to not add headers.
	// If GetProxyConnectHeader is non-nil, ProxyConnectHeader is
	// ignored.
	GetProxyConnectHeader() func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error)

	// MaxResponseHeaderBytes specifies a limit on how many
	// response bytes are allowed in the server's response
	// header.
	//
	// Zero means to use a default limit.
	MaxResponseHeaderBytes() int64

	// WriteBufferSize specifies the size of the write buffer used
	// when writing to the transport.
	// If zero, a default (currently 4KB) is used.
	WriteBufferSize() int

	// ReadBufferSize specifies the size of the read buffer used
	// when reading from the transport.
	// If zero, a default (currently 4KB) is used.
	ReadBufferSize() int

	TLSNextProto() map[string]func(authority string, c reqtls.Conn) http.RoundTripper

	SetTLSNextProto(map[string]func(authority string, c reqtls.Conn) http.RoundTripper)
}
