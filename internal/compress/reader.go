package compress

import "io"

type CompressReader interface {
	io.ReadCloser
	GetUnderlyingBody() io.ReadCloser
	SetUnderlyingBody(body io.ReadCloser)
}

func NewCompressReader(body io.ReadCloser, contentEncoding string) CompressReader {
	switch contentEncoding {
	case "gzip":
		return NewGzipReader(body)
	case "deflate":
		return NewDeflateReader(body)
	case "br":
		return NewBrotliReader(body)
	case "zstd":
		return NewZstdReader(body)
	}
	return nil
}
