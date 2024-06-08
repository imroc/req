package http3

import (
	"github.com/andybalholm/brotli"
	"io"
)

type BrotliReader struct {
	Body io.ReadCloser // underlying Response.Body
	br   io.Reader     // lazily-initialized brotli reader
	berr error         // sticky error
}

func newBrotliReader(body io.ReadCloser) io.ReadCloser {
	return &BrotliReader{Body: body}
}

func (br *BrotliReader) Read(p []byte) (n int, err error) {
	if br.berr != nil {
		return 0, br.berr
	}
	if br.br == nil {
		br.br = brotli.NewReader(br.Body)
	}
	return br.br.Read(p)
}

func (br *BrotliReader) Close() error {
	return br.Body.Close()
}
