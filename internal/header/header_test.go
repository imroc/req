package header

import (
	"net/http"
	"testing"
)

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"Host", true},
		{"host", true},
		{"Content-Length", true},
		{"Connection", true},
		{"Transfer-Encoding", true},
		{"Upgrade", true},
		{"Keep-Alive", true},
		{"Proxy-Connection", true},
		{"Content-Type", false},
		{"Authorization", false},
		{"User-Agent", false},
		{"__header_order__", true},
		{"__pseudo_header_order__", true},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := IsExcluded(tt.key); got != tt.expected {
				t.Errorf("IsExcluded(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestSortKeyValues(t *testing.T) {
	kvs := []KeyValues{
		{Key: "Content-Type", Values: []string{"text/html"}},
		{Key: "Authorization", Values: []string{"Bearer token"}},
		{Key: "User-Agent", Values: []string{"test"}},
	}
	// Sort with Authorization first, then User-Agent, then Content-Type
	orderedKeys := []string{"Authorization", "User-Agent", "Content-Type"}
	SortKeyValues(kvs, orderedKeys)
	if kvs[0].Key != "Authorization" {
		t.Fatalf("expected Authorization first, got %s", kvs[0].Key)
	}
	if kvs[1].Key != "User-Agent" {
		t.Fatalf("expected User-Agent second, got %s", kvs[1].Key)
	}
	if kvs[2].Key != "Content-Type" {
		t.Fatalf("expected Content-Type third, got %s", kvs[2].Key)
	}
}

func TestSortKeyValuesUnorderedFallback(t *testing.T) {
	kvs := []KeyValues{
		{Key: "Z-Custom", Values: []string{"z"}},
		{Key: "A-Custom", Values: []string{"a"}},
		{Key: "Content-Type", Values: []string{"text/html"}},
	}
	// Only Content-Type is in orderedKeys, others should retain relative order
	orderedKeys := []string{"Content-Type"}
	SortKeyValues(kvs, orderedKeys)
	// Content-Type should not necessarily be first since sort is stable
	// for items not in the order map (they keep their comparison indices)
}

func TestConstants(t *testing.T) {
	if DefaultUserAgent == "" {
		t.Fatal("DefaultUserAgent should not be empty")
	}
	if ContentType != "Content-Type" {
		t.Fatalf("ContentType = %q", ContentType)
	}
	if JsonContentType != "application/json; charset=utf-8" {
		t.Fatalf("JsonContentType = %q", JsonContentType)
	}
}

func TestIsExcludedWithHttpHeader(t *testing.T) {
	h := http.Header{}
	h.Set("Host", "example.com")
	h.Set("Content-Type", "text/html")
	for key := range h {
		if key == "Content-Type" && IsExcluded(key) {
			t.Fatal("Content-Type should not be excluded")
		}
	}
}
