package req

import (
	"bytes"
	"io"
	"testing"

	"github.com/imroc/req/v3/internal/tests"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestPeekDrain(t *testing.T) {
	a := autoDecodeReadCloser{peek: []byte("test")}
	p := make([]byte, 2)
	n, _ := a.peekDrain(p)
	tests.AssertEqual(t, 2, n)
	tests.AssertEqual(t, true, a.peek != nil)
	n, _ = a.peekDrain(p)
	tests.AssertEqual(t, 2, n)
	tests.AssertEqual(t, true, a.peek == nil)
}

func TestAutoDecodeReadCloserSplitMultibyte(t *testing.T) {
	gbkChina, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte("中国"))
	if err != nil {
		t.Fatal(err)
	}

	// First io.ReadAll read is 512 bytes. Place 中国 so 国 straddles that boundary:
	// 中 occupies 509-510, 国 occupies 511-512.
	const firstRead = 512
	const titleStart = 509
	head := []byte(`<!doctype html><html><head><meta charset="gbk"></head><body><title>`)
	if titleStart < len(head) {
		t.Fatalf("titleStart %d < head %d", titleStart, len(head))
	}
	body := bytes.Repeat([]byte{'x'}, titleStart)
	copy(body, head)
	body = append(body, gbkChina...)
	body = append(body, []byte(`</title></body></html>`)...)
	if body[firstRead-1] != gbkChina[2] || body[firstRead] != gbkChina[3] {
		t.Fatalf("国 is not split across the first-read boundary")
	}

	a := newAutoDecodeReadCloser(io.NopCloser(bytes.NewReader(body)), &Transport{})
	got, err := io.ReadAll(a)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("中国")) {
		t.Fatalf("decoded body missing 中国, got %q", got)
	}
}
