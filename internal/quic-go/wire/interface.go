package wire

import (
	"bytes"

	"github.com/imroc/req/v3/internal/quic-go/protocol"
)

// A Frame in QUIC
type Frame interface {
	Write(b *bytes.Buffer, version protocol.VersionNumber) error
	Length(version protocol.VersionNumber) protocol.ByteCount
}

// A FrameParser parses QUIC frames, one by one.
type FrameParser interface {
	ParseNext(*bytes.Reader, protocol.EncryptionLevel) (Frame, error)
	SetAckDelayExponent(uint8)
}
