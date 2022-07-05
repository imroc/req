package altsvcutil

import (
	"github.com/imroc/req/v3/internal/tests"
	"testing"
)

func TestParseHeader(t *testing.T) {
	as, err := ParseHeader(` h3=":443"; ma=86400, h3-29=":443"; ma=86400`)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 2, len(as))
	tests.AssertEqual(t, "h3", as[0].Protocol)
	tests.AssertEqual(t, "443", as[0].Port)
}
