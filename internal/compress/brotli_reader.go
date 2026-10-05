package compress

import (
	"io"

	"github.com/molecule-man/go-brrr"
)

type BrotliReader struct {
	Body io.ReadCloser // underlying Response.Body
	br   io.Reader     // lazily-initialized brotli reader
	berr error         // sticky error
}

func NewBrotliReader(body io.ReadCloser) *BrotliReader {
	return &BrotliReader{Body: body}
}

func (br *BrotliReader) Read(p []byte) (n int, err error) {
	if br.berr != nil {
		return 0, br.berr
	}
	if br.br == nil {
		br.br = brrr.NewReader(br.Body)
	}
	return br.br.Read(p)
}

func (br *BrotliReader) Close() error {
	return br.Body.Close()
}

func (br *BrotliReader) GetUnderlyingBody() io.ReadCloser {
	return br.Body
}

func (br *BrotliReader) SetUnderlyingBody(body io.ReadCloser) {
	br.Body = body
}
