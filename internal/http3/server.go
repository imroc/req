package http3

import "github.com/quic-go/quic-go"

// NextProtoH3 is the ALPN protocol negotiated during the TLS handshake, for QUIC v1 and v2.
const NextProtoH3 = "h3"

// StreamType is the stream type of a unidirectional stream.
type ServerStreamType uint64

const (
	streamTypeControlStream      = 0
	streamTypePushStream         = 1
	streamTypeQPACKEncoderStream = 2
	streamTypeQPACKDecoderStream = 3
)

func versionToALPN(v quic.Version) string {
	//nolint:exhaustive // These are all the versions we care about.
	switch v {
	case Version1, Version2:
		return NextProtoH3
	default:
		return ""
	}
}
