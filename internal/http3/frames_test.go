package http3

import (
	"bytes"
	"testing"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

func TestFrameParserReservedFrameType(t *testing.T) {
	for _, ft := range []uint64{0x2, 0x6, 0x8, 0x9} {
		data := quicvarint.Append(nil, ft)
		data = quicvarint.Append(data, 6)
		data = append(data, []byte("foobar")...)

		var closed bool
		fp := frameParser{
			streamID: 42,
			r:        bytes.NewReader(data),
			closeConn: func(quic.ApplicationErrorCode, string) error {
				closed = true
				return nil
			},
		}
		_, err := fp.ParseNext(nil)
		if err == nil {
			t.Fatalf("expected error for reserved frame type %d", ft)
		}
		if !closed {
			t.Fatalf("expected connection to be closed for reserved frame type %d", ft)
		}
	}
}

func TestFrameParserUnknownFrameType(t *testing.T) {
	data := quicvarint.Append(nil, 0xdead)
	data = quicvarint.Append(data, 6)
	data = append(data, []byte("foobar")...)
	// append a second frame (DATA) to verify parsing continues after unknown frame
	data = quicvarint.Append(data, 0x0) // DATA frame
	data = quicvarint.Append(data, 5)
	data = append(data, []byte("hello")...)

	fp := frameParser{
		streamID: 1,
		r:        bytes.NewReader(data),
		closeConn: func(quic.ApplicationErrorCode, string) error {
			return nil
		},
	}
	// First call should skip the unknown frame and parse the DATA frame
	f, err := fp.ParseNext(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	df, ok := f.(*dataFrame)
	if !ok {
		t.Fatalf("expected dataFrame, got %T", f)
	}
	if df.Length != 5 {
		t.Fatalf("expected length 5, got %d", df.Length)
	}
}

func TestFrameParserEOF(t *testing.T) {
	data := quicvarint.Append(nil, 0x0) // DATA frame
	data = quicvarint.Append(data, 6)
	data = append(data, []byte("foobar")...)

	// Truncate before the payload — should get an error (EOF or similar)
	// At length 0 and 1, varint reading fails with EOF
	for i := 0; i <= 1; i++ {
		b := make([]byte, i)
		copy(b, data[:i])
		fp := frameParser{r: bytes.NewReader(b)}
		_, err := fp.ParseNext(nil)
		if err == nil {
			t.Fatalf("expected error for truncated data at length %d", i)
		}
	}
}

func TestDataFrameAppend(t *testing.T) {
	f := &dataFrame{Length: 42}
	b := f.Append(nil)
	// frame type (0x0, 1 byte) + length (42, 1 byte) = 2 bytes
	if len(b) != 2 {
		t.Fatalf("expected 2 bytes, got %d", len(b))
	}
}

func TestHeadersFrameAppend(t *testing.T) {
	f := &headersFrame{Length: 100}
	b := f.Append(nil)
	// frame type (0x1, 1 byte) + length (100, 2 bytes since >63) = 3 bytes
	if len(b) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(b))
	}
}

func TestSettingsFrameAppendParse(t *testing.T) {
	sf := &settingsFrame{
		MaxFieldSectionSize: 1024,
		Datagram:            true,
		ExtendedConnect:     true,
		Other:               map[uint64]uint64{0x99: 1},
	}
	b := sf.Append(nil)

	// Parse it back
	r := &countingByteReader{Reader: quicvarint.NewReader(bytes.NewReader(b))}
	parsed, err := parseSettingsFrame(r, uint64(len(b)), 0, nil)
	if err != nil {
		t.Fatalf("failed to parse settings frame: %v", err)
	}
	if parsed.MaxFieldSectionSize != 1024 {
		t.Fatalf("expected MaxFieldSectionSize 1024, got %d", parsed.MaxFieldSectionSize)
	}
	if !parsed.Datagram {
		t.Fatal("expected Datagram to be true")
	}
	if !parsed.ExtendedConnect {
		t.Fatal("expected ExtendedConnect to be true")
	}
	if parsed.Other[0x99] != 1 {
		t.Fatalf("expected Other[0x99]=1, got %d", parsed.Other[0x99])
	}
}

func TestGoAwayFrameAppendParse(t *testing.T) {
	f := &goAwayFrame{StreamID: 42}
	b := f.Append(nil)

	// Skip the frame type (0x7) to get to the length + payload
	r := &countingByteReader{Reader: quicvarint.NewReader(bytes.NewReader(b))}
	// Read frame type
	_, err := quicvarint.Read(r)
	if err != nil {
		t.Fatalf("failed to read frame type: %v", err)
	}
	// Read length
	l, err := quicvarint.Read(r)
	if err != nil {
		t.Fatalf("failed to read length: %v", err)
	}
	parsed, err := parseGoAwayFrame(r, l, 0, nil)
	if err != nil {
		t.Fatalf("failed to parse goaway frame: %v", err)
	}
	if parsed.StreamID != 42 {
		t.Fatalf("expected StreamID 42, got %d", parsed.StreamID)
	}
}
