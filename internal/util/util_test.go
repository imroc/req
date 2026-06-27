package util

import (
	"os"
	"testing"
)

func TestIsJSONType(t *testing.T) {
	tests := []struct {
		ct       string
		expected bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"text/json", true},
		{"application/xml", false},
		{"text/html", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsJSONType(tt.ct); got != tt.expected {
			t.Errorf("IsJSONType(%q) = %v, want %v", tt.ct, got, tt.expected)
		}
	}
}

func TestIsXMLType(t *testing.T) {
	tests := []struct {
		ct       string
		expected bool
	}{
		{"application/xml", true},
		{"text/xml; charset=utf-8", true},
		{"application/json", false},
		{"text/html", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsXMLType(tt.ct); got != tt.expected {
			t.Errorf("IsXMLType(%q) = %v, want %v", tt.ct, got, tt.expected)
		}
	}
}

func TestIsStringEmpty(t *testing.T) {
	tests := []struct {
		s        string
		expected bool
	}{
		{"", true},
		{"  ", true},
		{"\t\n", true},
		{"a", false},
		{"  a  ", false},
	}
	for _, tt := range tests {
		if got := IsStringEmpty(tt.s); got != tt.expected {
			t.Errorf("IsStringEmpty(%q) = %v, want %v", tt.s, got, tt.expected)
		}
	}
}

func TestCutString(t *testing.T) {
	tests := []struct {
		s, sep   string
		before   string
		after    string
		found    bool
	}{
		{"a/b/c", "/", "a", "b/c", true},
		{"abc", "/", "abc", "", false},
		// Empty separator: strings.Index returns 0, so it "finds" at position 0
		{"abc", "", "", "abc", true},
	}
	for _, tt := range tests {
		before, after, found := CutString(tt.s, tt.sep)
		if before != tt.before || after != tt.after || found != tt.found {
			t.Errorf("CutString(%q, %q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.s, tt.sep, before, after, found, tt.before, tt.after, tt.found)
		}
	}
}

func TestCutBytes(t *testing.T) {
	tests := []struct {
		s, sep  []byte
		before  []byte
		after   []byte
		found   bool
	}{
		{[]byte("a/b/c"), []byte("/"), []byte("a"), []byte("b/c"), true},
		{[]byte("abc"), []byte("/"), []byte("abc"), nil, false},
	}
	for _, tt := range tests {
		before, after, found := CutBytes(tt.s, tt.sep)
		if string(before) != string(tt.before) || string(after) != string(tt.after) || found != tt.found {
			t.Errorf("CutBytes(%q, %q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.s, tt.sep, before, after, found, tt.before, tt.after, tt.found)
		}
	}
}

func TestBasicAuthHeaderValue(t *testing.T) {
	got := BasicAuthHeaderValue("user", "pass")
	// Basic dXNlcjpwYXNz
	if got != "Basic dXNlcjpwYXNz" {
		t.Fatalf("BasicAuthHeaderValue = %q", got)
	}
}

func TestGetPointer(t *testing.T) {
	type Foo struct{ X int }
	var f Foo
	p := GetPointer(&f)
	if p == nil {
		t.Fatal("GetPointer returned nil")
	}
	// p should be *Foo
	if _, ok := p.(*Foo); !ok {
		t.Fatalf("GetPointer returned %T, want *Foo", p)
	}

	// Non-pointer input
	p2 := GetPointer(Foo{})
	if p2 == nil {
		t.Fatal("GetPointer returned nil for non-pointer")
	}
}

func TestGetType(t *testing.T) {
	type Foo struct{ X int }
	f := &Foo{}
	if got := GetType(f); got.Name() != "Foo" {
		t.Fatalf("GetType = %s, want Foo", got.Name())
	}
}

func TestCreateDirectory(t *testing.T) {
	dir := t.TempDir() + "/testdir"
	if err := CreateDirectory(dir); err != nil {
		t.Fatalf("CreateDirectory failed: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("not a directory")
	}
	// Creating again should not error
	if err := CreateDirectory(dir); err != nil {
		t.Fatalf("CreateDirectory on existing dir failed: %v", err)
	}
}
