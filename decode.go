package req

import (
	"bytes"
	"io"
	"strings"

	"github.com/imroc/req/v3/internal/charsets"
)

var textContentTypes = []string{"text", "json", "xml", "html", "java"}

var autoDecodeText = autoDecodeContentTypeFunc(textContentTypes...)

func autoDecodeContentTypeFunc(contentTypes ...string) func(contentType string) bool {
	return func(contentType string) bool {
		for _, ct := range contentTypes {
			if strings.Contains(contentType, ct) {
				return true
			}
		}
		return false
	}
}

type decodeReaderCloser struct {
	io.ReadCloser
	decodeReader io.Reader
}

func (d *decodeReaderCloser) Read(p []byte) (n int, err error) {
	return d.decodeReader.Read(p)
}

func newAutoDecodeReadCloser(input io.ReadCloser, t *Transport) *autoDecodeReadCloser {
	return &autoDecodeReadCloser{ReadCloser: input, t: t}
}

type autoDecodeReadCloser struct {
	io.ReadCloser
	t            *Transport
	decodeReader io.Reader
	detected     bool
	peek         []byte
}

func (a *autoDecodeReadCloser) peekRead(p []byte) (n int, err error) {
	n, err = a.ReadCloser.Read(p)
	if n == 0 || (err != nil && err != io.EOF) {
		return
	}
	a.detected = true
	enc, name := charsets.FindEncoding(p[:n])
	if enc == nil {
		return
	}
	if a.t.Debugf != nil {
		a.t.Debugf("charset %s found in body's meta, auto-decode to utf-8", name)
	}
	// Replay the first chunk through the same decoder as the rest of the
	// body. Decoding it with Decoder.Bytes treats the chunk as a complete
	// input, so a multi-byte character split across this read is corrupted.
	first := make([]byte, n)
	copy(first, p[:n])
	a.decodeReader = enc.NewDecoder().Reader(io.MultiReader(bytes.NewReader(first), a.ReadCloser))
	return a.decodeReader.Read(p)
}

func (a *autoDecodeReadCloser) peekDrain(p []byte) (n int, err error) {
	if len(a.peek) > len(p) {
		copy(p, a.peek[:len(p)])
		peek := make([]byte, len(a.peek)-len(p))
		copy(peek, a.peek[len(p):])
		a.peek = peek
		n = len(p)
		return
	}
	if len(a.peek) == len(p) {
		copy(p, a.peek)
		n = len(p)
		a.peek = nil
		return
	}
	pp := make([]byte, len(p)-len(a.peek))
	nn, err := a.decodeReader.Read(pp)
	n = len(a.peek) + nn
	copy(p[:len(a.peek)], a.peek)
	copy(p[len(a.peek):], pp[:nn])
	a.peek = nil
	return
}

func (a *autoDecodeReadCloser) Read(p []byte) (n int, err error) {
	if !a.detected {
		return a.peekRead(p)
	}
	if a.peek != nil {
		return a.peekDrain(p)
	}
	if a.decodeReader != nil {
		return a.decodeReader.Read(p)
	}
	return a.ReadCloser.Read(p) // can not determine charset, not decode
}
