package wire

import (
	"bytes"
	"github.com/lucas-clemente/quic-go"

	"github.com/imroc/req/v3/internal/protocol"
	"github.com/imroc/req/v3/internal/quicvarint"
)

// A DataBlockedFrame is a DATA_BLOCKED frame
type DataBlockedFrame struct {
	MaximumData protocol.ByteCount
}

func parseDataBlockedFrame(r *bytes.Reader, _ quic.VersionNumber) (*DataBlockedFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}
	offset, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}
	return &DataBlockedFrame{
		MaximumData: protocol.ByteCount(offset),
	}, nil
}

func (f *DataBlockedFrame) Write(b *bytes.Buffer, version quic.VersionNumber) error {
	typeByte := uint8(0x14)
	b.WriteByte(typeByte)
	quicvarint.Write(b, uint64(f.MaximumData))
	return nil
}

// Length of a written frame
func (f *DataBlockedFrame) Length(version quic.VersionNumber) protocol.ByteCount {
	return 1 + protocol.ByteCount(quicvarint.Len(uint64(f.MaximumData)))
}
