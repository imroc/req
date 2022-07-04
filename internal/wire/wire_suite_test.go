package wire

import (
	"bytes"
	"encoding/binary"
	"github.com/lucas-clemente/quic-go"
	"testing"

	"github.com/imroc/req/v3/internal/quicvarint"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWire(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Wire Suite")
}

func encodeVarInt(i uint64) []byte {
	b := &bytes.Buffer{}
	quicvarint.Write(b, i)
	return b.Bytes()
}

func appendVersion(data []byte, v quic.VersionNumber) []byte {
	offset := len(data)
	data = append(data, []byte{0, 0, 0, 0}...)
	binary.BigEndian.PutUint32(data[offset:], uint32(v))
	return data
}
