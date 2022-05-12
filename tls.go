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
	// ConnectionState returns basic TLS details about the connection.
	ConnectionState() tls.ConnectionState
	// Handshake runs the client or server handshake
	// protocol if it has not yet been run.
	//
	// Most uses of this package need not call Handshake explicitly: the
	// first Read or Write will call it automatically.
	//
	// For control over canceling or setting a timeout on a handshake, use
	// HandshakeContext or the Dialer's DialContext method instead.
	Handshake() error
}

// NetConnWrapper is the interface to get underlying connection, which is
// introduced in go1.18 for *tls.Conn.
type NetConnWrapper interface {
	// NetConn returns the underlying connection that is wrapped by c.
	// Note that writing to or reading from this connection directly will corrupt the
	// TLS session.
	NetConn() net.Conn
}
