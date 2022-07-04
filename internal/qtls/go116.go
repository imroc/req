//go:build go1.16 && !go1.17
// +build go1.16,!go1.17

package qtls

import (
	"crypto/tls"
	"github.com/marten-seemann/qtls-go1-16"
)

type (
	// ConnectionState contains information about the state of the connection.
	ConnectionState = qtls.ConnectionStateWith0RTT
)

// ToTLSConnectionState extracts the tls.ConnectionState
func ToTLSConnectionState(cs ConnectionState) tls.ConnectionState {
	return cs.ConnectionState
}
