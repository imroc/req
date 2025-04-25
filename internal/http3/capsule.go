package http3

import "io"

type countingByteReader struct {
	io.ByteReader
	Read int
}

func (r *countingByteReader) ReadByte() (byte, error) {
	b, err := r.ByteReader.ReadByte()
	if err == nil {
		r.Read++
	}
	return b, err
}
