package tlsclient

import (
	"crypto/tls"
	"net"
)

type Conn interface {
	net.Conn
	ConnectionState() tls.ConnectionState
	Handshake() error
}
