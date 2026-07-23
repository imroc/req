package req

import (
	"testing"
	"time"
)

func TestSetTimeoutNegativeIsZero(t *testing.T) {
	c := C()
	c.SetTimeout(-time.Second)
	if c.GetClient().Timeout != 0 {
		t.Fatalf("got %v want 0", c.GetClient().Timeout)
	}
	c.SetTimeout(5 * time.Second)
	if c.GetClient().Timeout != 5*time.Second {
		t.Fatalf("got %v", c.GetClient().Timeout)
	}
}
