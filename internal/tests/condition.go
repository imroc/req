package tests

import "time"

// WaitCondition reports whether fn eventually returned true,
// checking immediately and then every checkEvery amount,
// until waitFor has elapsed, at which point it returns false.
func WaitCondition(waitFor, checkEvery time.Duration, fn func() bool) bool {
	deadline := time.Now().Add(waitFor)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(checkEvery)
	}
	return false
}
