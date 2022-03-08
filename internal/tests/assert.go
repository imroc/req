package tests

import (
	"reflect"
	"strings"
	"testing"
)

// AssertIsNil asserts is nil.
func AssertIsNil(t *testing.T, v interface{}) {
	if !isNil(v) {
		t.Errorf("[%v] was expected to be nil", v)
	}
}

// AssertNotNil asserts is not nil.
func AssertNotNil(t *testing.T, v interface{}) {
	if isNil(v) {
		t.Fatalf("[%v] was expected to be non-nil", v)
	}
}

// AssertEqual asserts e (expected) is equal with g (got).
func AssertEqual(t *testing.T, e, g interface{}) {
	if !equal(e, g) {
		t.Errorf("Expected [%+v], got [%+v]", e, g)
	}
	return
}

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

func equal(expected, got interface{}) bool {
	return reflect.DeepEqual(expected, got)
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && rv.IsNil() {
		return true
	}
	return false
}
