package req

import (
	"crypto/tls"
	"net"
)

// TLSConn is the recommended interface for the connection
// returned by the DailTLS function (Client.SetDialTLS,
// Transport.DialTLSContext), so that the TLS handshake negotiation
// can automatically decide whether to use HTTP2 or HTTP1 (ALPN).
// If this interface is not implemented, HTTP1 will be used by default.
type TLSConn interface {
	net.Conn
	ConnectionState() tls.ConnectionState
	Handshake() error
}
