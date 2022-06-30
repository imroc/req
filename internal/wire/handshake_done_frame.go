package wire

import (
	"bytes"
	"github.com/lucas-clemente/quic-go"

	"github.com/imroc/req/v3/internal/protocol"
)

// A HandshakeDoneFrame is a HANDSHAKE_DONE frame
type HandshakeDoneFrame struct{}

// ParseHandshakeDoneFrame parses a HandshakeDone frame
func parseHandshakeDoneFrame(r *bytes.Reader, _ quic.VersionNumber) (*HandshakeDoneFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}
	return &HandshakeDoneFrame{}, nil
}

func (f *HandshakeDoneFrame) Write(b *bytes.Buffer, _ quic.VersionNumber) error {
	b.WriteByte(0x1e)
	return nil
}

// Length of a written frame
func (f *HandshakeDoneFrame) Length(_ quic.VersionNumber) protocol.ByteCount {
	return 1
}
