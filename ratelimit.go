package req

import (
	"io"
	"time"
)

// rateLimitedReadCloser wraps an io.ReadCloser and paces reads so that the
// average throughput does not exceed limit bytes per second. It is used for
// both upload (the transport reads the request body through it) and download
// (the response body is read through it) bandwidth limiting.
//
// After each read it sleeps for however long that many bytes should have taken
// at the configured rate, minus the time the read itself already took. Data is
// therefore never delayed longer than necessary and the average rate converges
// to the limit, at the cost of allowing a single read (at most one buffer) to
// burst through before the pause.
type rateLimitedReadCloser struct {
	rc    io.ReadCloser
	limit float64 // bytes per second, always > 0
}

func newRateLimitedReadCloser(rc io.ReadCloser, bytesPerSecond int64) *rateLimitedReadCloser {
	return &rateLimitedReadCloser{
		rc:    rc,
		limit: float64(bytesPerSecond),
	}
}

func (l *rateLimitedReadCloser) Read(p []byte) (int, error) {
	start := time.Now()
	n, err := l.rc.Read(p)
	if n > 0 {
		expected := time.Duration(float64(n) / l.limit * float64(time.Second))
		if d := expected - time.Since(start); d > 0 {
			time.Sleep(d)
		}
	}
	return n, err
}

func (l *rateLimitedReadCloser) Close() error {
	return l.rc.Close()
}
