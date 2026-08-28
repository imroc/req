package req

import (
	"math"
	"testing"
)

func TestSetHTTP2NextStreamID(t *testing.T) {
	c := C()

	// odd value passes through to the internal h2 transport
	c.SetHTTP2NextStreamID(3)
	if got := c.Transport.t2.NextStreamID; got != 3 {
		t.Errorf("t2.NextStreamID = %d, want 3", got)
	}

	// even values are ignored (client stream IDs must be odd)
	c.SetHTTP2NextStreamID(4)
	if got := c.Transport.t2.NextStreamID; got != 3 {
		t.Errorf("t2.NextStreamID = %d, want 3 (even value ignored)", got)
	}

	// values beyond 31 bits are ignored (RFC 9113)
	c.SetHTTP2NextStreamID(math.MaxInt32 + 2)
	if got := c.Transport.t2.NextStreamID; got != 3 {
		t.Errorf("t2.NextStreamID = %d, want 3 (>31-bit value ignored)", got)
	}

	// Transport.Clone carries the setting over
	clone := c.Transport.Clone()
	if got := clone.t2.NextStreamID; got != 3 {
		t.Errorf("cloned t2.NextStreamID = %d, want 3", got)
	}
}
