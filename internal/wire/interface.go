package wire

import (
	"bytes"
	"github.com/lucas-clemente/quic-go"

	"github.com/imroc/req/v3/internal/protocol"
)

// A Frame in QUIC
type Frame interface {
	Write(b *bytes.Buffer, version quic.VersionNumber) error
	Length(version quic.VersionNumber) protocol.ByteCount
}

// A FrameParser parses QUIC frames, one by one.
type FrameParser interface {
	ParseNext(*bytes.Reader, protocol.EncryptionLevel) (Frame, error)
	SetAckDelayExponent(uint8)
}
