package req

import (
	"bytes"
	"log"
	"testing"

	"github.com/0xobjc/req/v3/internal/tests"
)

func TestLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	l := NewLogger(buf, "", log.Ldate|log.Lmicroseconds)
	c := tc().SetLogger(l)
	c.SetProxyURL(":=\\<>ksfj&*&sf")
	tests.AssertContains(t, buf.String(), "error", true)
	buf.Reset()
	c.R().SetOutput(nil)
	tests.AssertContains(t, buf.String(), "warn", true)
}

func TestLoggerFromStandardLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	l := NewLoggerFromStandardLogger(log.New(buf, "", log.Ldate|log.Lmicroseconds))
	c := tc().SetLogger(l)
	c.SetProxyURL(":=\\<>ksfj&*&sf")
	tests.AssertContains(t, buf.String(), "error", true)
	buf.Reset()
	c.R().SetOutput(nil)
	tests.AssertContains(t, buf.String(), "warn", true)
}
