package protocol

import (
	"github.com/lucas-clemente/quic-go"
)

// The version numbers, making grepping easier
const (
	VersionTLS     quic.VersionNumber = 0x1
	VersionDraft29 quic.VersionNumber = 0xff00001d
	Version1       quic.VersionNumber = 0x1
	Version2       quic.VersionNumber = 0x709a50c4
)
