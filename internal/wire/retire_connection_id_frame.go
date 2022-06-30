package wire

import (
	"bytes"
	"github.com/lucas-clemente/quic-go"

	"github.com/imroc/req/v3/internal/protocol"
	"github.com/imroc/req/v3/internal/quicvarint"
)

// A RetireConnectionIDFrame is a RETIRE_CONNECTION_ID frame
type RetireConnectionIDFrame struct {
	SequenceNumber uint64
}

func parseRetireConnectionIDFrame(r *bytes.Reader, _ quic.VersionNumber) (*RetireConnectionIDFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	seq, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}
	return &RetireConnectionIDFrame{SequenceNumber: seq}, nil
}

func (f *RetireConnectionIDFrame) Write(b *bytes.Buffer, _ quic.VersionNumber) error {
	b.WriteByte(0x19)
	quicvarint.Write(b, f.SequenceNumber)
	return nil
}

// Length of a written frame
func (f *RetireConnectionIDFrame) Length(quic.VersionNumber) protocol.ByteCount {
	return 1 + protocol.ByteCount(quicvarint.Len(f.SequenceNumber))
}
