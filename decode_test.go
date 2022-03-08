package req

import (
	"testing"
)

func TestPeekDrain(t *testing.T) {
	a := autoDecodeReadCloser{peek: []byte("test")}
	p := make([]byte, 2)
	n, _ := a.peekDrain(p)
	assertEqual(t, 2, n)
	assertEqual(t, true, a.peek != nil)
	n, _ = a.peekDrain(p)
	assertEqual(t, 2, n)
	assertEqual(t, true, a.peek == nil)
}
