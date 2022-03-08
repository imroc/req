package req

import (
	"github.com/imroc/req/v3/internal/tests"
	"testing"
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
