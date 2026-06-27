package netutil

import (
	"net/url"
	"testing"
)

func TestAuthorityAddr(t *testing.T) {
	tests := []struct {
		scheme    string
		authority string
		expected  string
	}{
		{"https", "example.com", "example.com:443"},
		{"http", "example.com", "example.com:80"},
		{"https", "example.com:8443", "example.com:8443"},
		{"http", "example.com:8080", "example.com:8080"},
		{"https", "[::1]:443", "[::1]:443"},
		{"https", "[::1]", "[::1]:443"},
	}
	for _, tt := range tests {
		t.Run(tt.scheme+"://"+tt.authority, func(t *testing.T) {
			got := AuthorityAddr(tt.scheme, tt.authority)
			if got != tt.expected {
				t.Errorf("AuthorityAddr(%q, %q) = %q, want %q",
					tt.scheme, tt.authority, got, tt.expected)
			}
		})
	}
}

func TestAuthorityHostPort(t *testing.T) {
	tests := []struct {
		scheme    string
		authority string
		host      string
		port      string
	}{
		{"https", "example.com", "example.com", "443"},
		{"http", "example.com", "example.com", "80"},
		{"https", "example.com:8443", "example.com", "8443"},
	}
	for _, tt := range tests {
		host, port := AuthorityHostPort(tt.scheme, tt.authority)
		if host != tt.host || port != tt.port {
			t.Errorf("AuthorityHostPort(%q, %q) = (%q, %q), want (%q, %q)",
				tt.scheme, tt.authority, host, port, tt.host, tt.port)
		}
	}
}

func TestAuthorityKey(t *testing.T) {
	u := &url.URL{Scheme: "https", Host: "example.com"}
	got := AuthorityKey(u)
	if got != "https://example.com:443" {
		t.Fatalf("AuthorityKey = %q, want https://example.com:443", got)
	}
}
