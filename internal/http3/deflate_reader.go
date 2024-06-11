package http3

import (
	"compress/flate"
	"io"
)

type DeflateReader struct {
	Body io.ReadCloser // underlying Response.Body
	dr   io.ReadCloser // lazily-initialized deflate reader
	derr error         // sticky error
}

func newDeflateReader(body io.ReadCloser) io.ReadCloser {
	return &DeflateReader{Body: body}
}

func (df *DeflateReader) Read(p []byte) (n int, err error) {
	if df.derr != nil {
		return 0, df.derr
	}
	if df.dr == nil {
		df.dr = flate.NewReader(df.Body)
		if df.dr == nil {
			df.derr = io.ErrUnexpectedEOF
			return 0, df.derr
		}
	}
	return df.dr.Read(p)
}

func (df *DeflateReader) Close() error {
	if df.dr != nil {
		err := df.dr.Close()
		if err != nil {
			return err
		}
	}
	return df.Body.Close()
}
