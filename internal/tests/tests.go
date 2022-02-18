package tests

import (
	"strings"
	"testing"
)

// AssertNoError asserts no error.
func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

// AssertErrorContains asserts error is not nil and contains specified error string.
func AssertErrorContains(t *testing.T, err error, s string) {
	if err == nil {
		t.Error("err is nil")
		return
	}
	if !strings.Contains(err.Error(), s) {
		t.Errorf("%q is not included in error %q", s, err.Error())
	}
}

// AssertContains asserts substring is contained in the given string.
func AssertContains(t *testing.T, s, substr string, shouldContain bool) {
	s = strings.ToLower(s)
	isContain := strings.Contains(s, substr)
	if shouldContain {
		if !isContain {
			t.Errorf("%q is not included in %s", substr, s)
		}
	} else {
		if isContain {
			t.Errorf("%q is included in %s", substr, s)
		}
	}
}
