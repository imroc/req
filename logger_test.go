package req

import (
	"bytes"
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	l := NewLogger(buf, "", log.Ldate|log.Lmicroseconds)
	c := tc().SetLogger(l)
	c.SetProxyURL(":=\\<>ksfj&*&sf")
	assertContains(t, buf.String(), "error", true)
	buf.Reset()
	c.R().SetOutput(nil)
	assertContains(t, buf.String(), "warn", true)
}
