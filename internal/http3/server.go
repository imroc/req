package http3

// NextProtoH3 is the ALPN protocol negotiated during the TLS handshake, for QUIC v1 and v2.
const NextProtoH3 = "h3"

const (
	streamTypeControlStream      = 0
	streamTypePushStream         = 1
	streamTypeQPACKEncoderStream = 2
	streamTypeQPACKDecoderStream = 3
)
