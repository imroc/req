//go:build go1.19
// +build go1.19

package qtls

import (
	"crypto/tls"
	"github.com/marten-seemann/qtls-go1-19"
)

type (
	// ConnectionState contains information about the state of the connection.
	ConnectionState = qtls.ConnectionStateWith0RTT
)

// ToTLSConnectionState extracts the tls.ConnectionState
func ToTLSConnectionState(cs ConnectionState) tls.ConnectionState {
	return cs.ConnectionState
}
