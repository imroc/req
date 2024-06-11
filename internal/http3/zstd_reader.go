package http3

import (
	"github.com/klauspost/compress/zstd"
	"io"
)

type ZstdReader struct {
	Body io.ReadCloser // underlying Response.Body
	zr   *zstd.Decoder // lazily-initialized zstd reader
	zerr error         // sticky error
}

func newZstdReader(body io.ReadCloser) io.ReadCloser {
	return &ZstdReader{Body: body}
}

func (zr *ZstdReader) Read(p []byte) (n int, err error) {
	if zr.zerr != nil {
		return 0, zr.zerr
	}
	if zr.zr == nil {
		zr.zr, err = zstd.NewReader(zr.Body)
		if err != nil {
			zr.zerr = err
			return 0, err
		}
	}
	return zr.zr.Read(p)
}

func (zr *ZstdReader) Close() error {
	if zr.zr != nil {
		zr.zr.Close()
	}
	return zr.Body.Close()
}
