package wire

import (
	"bytes"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"io"

	"github.com/imroc/req/v3/internal/protocol"
	"github.com/imroc/req/v3/internal/quicvarint"
)

// A NewConnectionIDFrame is a NEW_CONNECTION_ID frame
type NewConnectionIDFrame struct {
	SequenceNumber      uint64
	RetirePriorTo       uint64
	ConnectionID        protocol.ConnectionID
	StatelessResetToken protocol.StatelessResetToken
}

func parseNewConnectionIDFrame(r *bytes.Reader, _ quic.VersionNumber) (*NewConnectionIDFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	seq, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}
	ret, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}
	if ret > seq {
		//nolint:stylecheck
		return nil, fmt.Errorf("Retire Prior To value (%d) larger than Sequence Number (%d)", ret, seq)
	}
	connIDLen, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if connIDLen > protocol.MaxConnIDLen {
		return nil, fmt.Errorf("invalid connection ID length: %d", connIDLen)
	}
	connID, err := protocol.ReadConnectionID(r, int(connIDLen))
	if err != nil {
		return nil, err
	}
	frame := &NewConnectionIDFrame{
		SequenceNumber: seq,
		RetirePriorTo:  ret,
		ConnectionID:   connID,
	}
	if _, err := io.ReadFull(r, frame.StatelessResetToken[:]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, err
	}

	return frame, nil
}

func (f *NewConnectionIDFrame) Write(b *bytes.Buffer, _ quic.VersionNumber) error {
	b.WriteByte(0x18)
	quicvarint.Write(b, f.SequenceNumber)
	quicvarint.Write(b, f.RetirePriorTo)
	connIDLen := f.ConnectionID.Len()
	if connIDLen > protocol.MaxConnIDLen {
		return fmt.Errorf("invalid connection ID length: %d", connIDLen)
	}
	b.WriteByte(uint8(connIDLen))
	b.Write(f.ConnectionID.Bytes())
	b.Write(f.StatelessResetToken[:])
	return nil
}

// Length of a written frame
func (f *NewConnectionIDFrame) Length(quic.VersionNumber) protocol.ByteCount {
	return 1 + protocol.ByteCount(quicvarint.Len(f.SequenceNumber)+quicvarint.Len(f.RetirePriorTo)) + 1 /* connection ID length */ + protocol.ByteCount(f.ConnectionID.Len()) + 16
}
