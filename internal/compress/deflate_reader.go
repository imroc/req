package compress

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"io"
)

type DeflateReader struct {
	Body io.ReadCloser // underlying Response.Body
	dr   io.ReadCloser // lazily-initialized deflate reader
	derr error         // sticky error
}

func NewDeflateReader(body io.ReadCloser) *DeflateReader {
	return &DeflateReader{Body: body}
}

// looksLikeZlib reports whether the two bytes are an RFC 1950 zlib header:
// the compression method is deflate, the window size is in range, and the
// pair is a multiple of 31. A raw deflate stream can satisfy all three by
// chance, so this is a heuristic, but it is the one browsers use and it is
// the only way to tell the two apart.
func looksLikeZlib(cmf, flg byte) bool {
	if cmf&0x0f != 8 { // CM, 8 is deflate
		return false
	}
	if cmf>>4 > 7 { // CINFO, window size
		return false
	}
	return (uint16(cmf)<<8|uint16(flg))%31 == 0 // FCHECK
}

func (df *DeflateReader) Read(p []byte) (n int, err error) {
	if df.derr != nil {
		return 0, df.derr
	}
	if df.dr == nil {
		// Content-Encoding: deflate is the zlib format of RFC 1950, but raw
		// RFC 1951 deflate is common enough that both have to be read. Sniff
		// the header and put the bytes back.
		var head [2]byte
		got, herr := io.ReadFull(df.Body, head[:])
		if herr != nil && herr != io.EOF && herr != io.ErrUnexpectedEOF {
			df.derr = herr
			return 0, herr
		}
		body := io.MultiReader(bytes.NewReader(head[:got]), df.Body)
		if got == len(head) && looksLikeZlib(head[0], head[1]) {
			zr, zerr := zlib.NewReader(body)
			if zerr != nil {
				df.derr = zerr
				return 0, zerr
			}
			df.dr = zr
		} else {
			df.dr = flate.NewReader(body)
		}
	}
	return df.dr.Read(p)
}

func (df *DeflateReader) Close() error {
	if df.dr != nil {
		df.dr.Close()
	}
	return df.Body.Close()
}

func (df *DeflateReader) GetUnderlyingBody() io.ReadCloser {
	return df.Body
}

func (df *DeflateReader) SetUnderlyingBody(body io.ReadCloser) {
	df.Body = body
}
