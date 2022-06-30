package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/lucas-clemente/quic-go"
	"math"
)

// gQUIC version range as defined in the wiki: https://github.com/quicwg/base-drafts/wiki/QUIC-Versions
const (
	gquicVersion0   = 0x51303030
	maxGquicVersion = 0x51303439
)

// The version numbers, making grepping easier
const (
	VersionTLS      quic.VersionNumber = 0x1
	VersionWhatever quic.VersionNumber = math.MaxUint32 - 1 // for when the version doesn't matter
	VersionUnknown  quic.VersionNumber = math.MaxUint32
	VersionDraft29  quic.VersionNumber = 0xff00001d
	Version1        quic.VersionNumber = 0x1
	Version2        quic.VersionNumber = 0x709a50c4
)

// SupportedVersions lists the versions that the server supports
// must be in sorted descending order
var SupportedVersions = []quic.VersionNumber{Version1, Version2, VersionDraft29}

// IsValidVersion says if the version is known to quic-go
func IsValidVersion(v quic.VersionNumber) bool {
	return v == VersionTLS || IsSupportedVersion(SupportedVersions, v)
}

// IsSupportedVersion returns true if the server supports this version
func IsSupportedVersion(supported []quic.VersionNumber, v quic.VersionNumber) bool {
	for _, t := range supported {
		if t == v {
			return true
		}
	}
	return false
}

// ChooseSupportedVersion finds the best version in the overlap of ours and theirs
// ours is a slice of versions that we support, sorted by our preference (descending)
// theirs is a slice of versions offered by the peer. The order does not matter.
// The bool returned indicates if a matching version was found.
func ChooseSupportedVersion(ours, theirs []quic.VersionNumber) (quic.VersionNumber, bool) {
	for _, ourVer := range ours {
		for _, theirVer := range theirs {
			if ourVer == theirVer {
				return ourVer, true
			}
		}
	}
	return 0, false
}

// generateReservedVersion generates a reserved version number (v & 0x0f0f0f0f == 0x0a0a0a0a)
func generateReservedVersion() quic.VersionNumber {
	b := make([]byte, 4)
	_, _ = rand.Read(b) // ignore the error here. Failure to read random data doesn't break anything
	return quic.VersionNumber((binary.BigEndian.Uint32(b) | 0x0a0a0a0a) & 0xfafafafa)
}

// GetGreasedVersions adds one reserved version number to a slice of version numbers, at a random position
func GetGreasedVersions(supported []quic.VersionNumber) []quic.VersionNumber {
	b := make([]byte, 1)
	_, _ = rand.Read(b) // ignore the error here. Failure to read random data doesn't break anything
	randPos := int(b[0]) % (len(supported) + 1)
	greased := make([]quic.VersionNumber, len(supported)+1)
	copy(greased, supported[:randPos])
	greased[randPos] = generateReservedVersion()
	copy(greased[randPos+1:], supported[randPos:])
	return greased
}
