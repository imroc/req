package http2

import (
	"io"
	"net"
	"testing"

	reqhttp2 "github.com/imroc/req/v3/http2"
)

// newPipeClientConn drives newClientConn over a net.Pipe whose peer merely
// drains the client's preface/SETTINGS/priority frames without sending any
// server response.
func newPipeClientConn(t *testing.T, tr *Transport, singleUse bool) *ClientConn {
	t.Helper()
	c1, c2 := net.Pipe()
	t.Cleanup(func() { c1.Close() })
	t.Cleanup(func() { c2.Close() })
	go io.Copy(io.Discard, c2)

	cc, err := tr.newClientConn(c1, singleUse)
	if err != nil {
		t.Fatalf("newClientConn: %v", err)
	}
	t.Cleanup(func() { cc.Close() })
	return cc
}

func TestNextStreamID(t *testing.T) {
	t.Run("default is 1", func(t *testing.T) {
		if got := newPipeClientConn(t, &Transport{}, false).nextStreamID; got != 1 {
			t.Errorf("nextStreamID = %d, want 1", got)
		}
	})

	t.Run("custom odd value applied", func(t *testing.T) {
		// OkHttp starts its first request stream at 3.
		tr := &Transport{NextStreamID: 3}
		if got := newPipeClientConn(t, tr, false).nextStreamID; got != 3 {
			t.Errorf("nextStreamID = %d, want 3", got)
		}
	})

	t.Run("even value ignored", func(t *testing.T) {
		// Client stream IDs must be odd (RFC 9113); even values are ignored.
		tr := &Transport{NextStreamID: 4}
		if got := newPipeClientConn(t, tr, false).nextStreamID; got != 1 {
			t.Errorf("nextStreamID = %d, want 1", got)
		}
	})

	t.Run("value beyond 31 bits ignored", func(t *testing.T) {
		// Stream IDs are 31-bit (RFC 9113); larger values are ignored.
		tr := &Transport{NextStreamID: 1<<31 + 1}
		if got := newPipeClientConn(t, tr, false).nextStreamID; got != 1 {
			t.Errorf("nextStreamID = %d, want 1", got)
		}
	})

	t.Run("priority frames advance past claimed streams", func(t *testing.T) {
		// Firefox-style priority tree: placeholder streams 3..13, so the
		// first request stream must be 15.
		tr := &Transport{
			PriorityFrames: []reqhttp2.PriorityFrame{
				{StreamID: 3, PriorityParam: reqhttp2.PriorityParam{Weight: 200}},
				{StreamID: 13, PriorityParam: reqhttp2.PriorityParam{Weight: 240}},
			},
		}
		if got := newPipeClientConn(t, tr, false).nextStreamID; got != 15 {
			t.Errorf("nextStreamID = %d, want 15", got)
		}
	})

	t.Run("custom base with priority frames still advances", func(t *testing.T) {
		tr := &Transport{
			NextStreamID: 5,
			PriorityFrames: []reqhttp2.PriorityFrame{
				{StreamID: 7, PriorityParam: reqhttp2.PriorityParam{Weight: 100}},
			},
		}
		if got := newPipeClientConn(t, tr, false).nextStreamID; got != 9 {
			t.Errorf("nextStreamID = %d, want 9", got)
		}
	})

	t.Run("singleUse conn stays usable with custom NextStreamID", func(t *testing.T) {
		// Regression test: a singleUse connection must accept its first
		// request even when the stream ID counter does not start at 1.
		cc := newPipeClientConn(t, &Transport{NextStreamID: 3}, true)
		if st := cc.idleState(); !st.canTakeNewRequest {
			t.Error("singleUse conn with NextStreamID=3 cannot take its first request")
		}
	})

	t.Run("singleUse conn rejects second request", func(t *testing.T) {
		cc := newPipeClientConn(t, &Transport{NextStreamID: 3}, true)
		cc.nextStreamID += 2 // simulate one request stream allocated
		if st := cc.idleState(); st.canTakeNewRequest {
			t.Error("singleUse conn unexpectedly accepts a second request")
		}
	})
}
