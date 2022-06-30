package tests

import (
	"go/token"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func AssertIsNil(t *testing.T, v interface{}) {
	if !isNil(v) {
		t.Errorf("[%v] was expected to be nil", v)
	}
}

func AssertAllNotNil(t *testing.T, vv ...interface{}) {
	for _, v := range vv {
		AssertNotNil(t, v)
	}
}

func AssertNotNil(t *testing.T, v interface{}) {
	if isNil(v) {
		t.Fatalf("[%v] was expected to be non-nil", v)
	}
}

func AssertEqual(t *testing.T, e, g interface{}) {
	if !equal(e, g) {
		t.Errorf("Expected [%+v], got [%+v]", e, g)
	}
	return
}

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func AssertErrorContains(t *testing.T, err error, s string) {
	if err == nil {
		t.Error("err is nil")
		return
	}
	if !strings.Contains(err.Error(), s) {
		t.Errorf("%q is not included in error %q", s, err.Error())
	}
}

func AssertContains(t *testing.T, s, substr string, shouldContain bool) {
	s = strings.ToLower(s)
	isContain := strings.Contains(s, substr)
	if shouldContain {
		if !isContain {
			t.Errorf("%q is not included in %s", substr, s)
		}
	} else {
		if isContain {
			t.Errorf("%q is included in %q", substr, s)
		}
	}
}

func AssertClone(t *testing.T, e, g interface{}) {
	ev := reflect.ValueOf(e).Elem()
	gv := reflect.ValueOf(g).Elem()
	et := ev.Type()

	for i := 0; i < ev.NumField(); i++ {
		sf := ev.Field(i)
		st := et.Field(i)

		var ee, gg interface{}
		if !token.IsExported(st.Name) {
			ee = reflect.NewAt(sf.Type(), unsafe.Pointer(sf.UnsafeAddr())).Elem().Interface()
			gg = reflect.NewAt(sf.Type(), unsafe.Pointer(gv.Field(i).UnsafeAddr())).Elem().Interface()
		} else {
			ee = sf.Interface()
			gg = gv.Field(i).Interface()
		}
		if sf.Kind() == reflect.Func || sf.Kind() == reflect.Slice || sf.Kind() == reflect.Ptr {
			if ee != nil {
				if gg == nil {
					t.Errorf("Field %s.%s is nil", et.Name(), et.Field(i).Name)
				}
			}
			continue
		}
		if !reflect.DeepEqual(ee, gg) {
			t.Errorf("Field %s.%s is not equal, expected [%v], got [%v]", et.Name(), et.Field(i).Name, ee, gg)
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
