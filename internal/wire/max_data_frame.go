package wire

import (
	"bytes"
	"github.com/lucas-clemente/quic-go"

	"github.com/imroc/req/v3/internal/protocol"
	"github.com/imroc/req/v3/internal/quicvarint"
)

// A MaxDataFrame carries flow control information for the connection
type MaxDataFrame struct {
	MaximumData protocol.ByteCount
}

// parseMaxDataFrame parses a MAX_DATA frame
func parseMaxDataFrame(r *bytes.Reader, _ quic.VersionNumber) (*MaxDataFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	frame := &MaxDataFrame{}
	byteOffset, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}
	frame.MaximumData = protocol.ByteCount(byteOffset)
	return frame, nil
}

// Write writes a MAX_STREAM_DATA frame
func (f *MaxDataFrame) Write(b *bytes.Buffer, version quic.VersionNumber) error {
	b.WriteByte(0x10)
	quicvarint.Write(b, uint64(f.MaximumData))
	return nil
}

// Length of a written frame
func (f *MaxDataFrame) Length(version quic.VersionNumber) protocol.ByteCount {
	return 1 + protocol.ByteCount(quicvarint.Len(uint64(f.MaximumData)))
}
