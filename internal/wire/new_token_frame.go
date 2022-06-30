package wire

import (
	"bytes"
	"errors"
	"github.com/lucas-clemente/quic-go"
	"io"

	"github.com/imroc/req/v3/internal/protocol"
	"github.com/imroc/req/v3/internal/quicvarint"
)

// A NewTokenFrame is a NEW_TOKEN frame
type NewTokenFrame struct {
	Token []byte
}

func parseNewTokenFrame(r *bytes.Reader, _ quic.VersionNumber) (*NewTokenFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}
	tokenLen, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}
	if uint64(r.Len()) < tokenLen {
		return nil, io.EOF
	}
	if tokenLen == 0 {
		return nil, errors.New("token must not be empty")
	}
	token := make([]byte, int(tokenLen))
	if _, err := io.ReadFull(r, token); err != nil {
		return nil, err
	}
	return &NewTokenFrame{Token: token}, nil
}

func (f *NewTokenFrame) Write(b *bytes.Buffer, _ quic.VersionNumber) error {
	b.WriteByte(0x7)
	quicvarint.Write(b, uint64(len(f.Token)))
	b.Write(f.Token)
	return nil
}

// Length of a written frame
func (f *NewTokenFrame) Length(quic.VersionNumber) protocol.ByteCount {
	return 1 + protocol.ByteCount(quicvarint.Len(uint64(len(f.Token)))) + protocol.ByteCount(len(f.Token))
}
