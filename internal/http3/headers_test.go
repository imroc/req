package http3

import (
	"io"
	"net/http"
	"testing"

	"github.com/imroc/req/v3/internal/dump"
	"github.com/quic-go/qpack"
)

func TestParseHeadersRejectsUserinfoInAuthority(t *testing.T) {
	fields := []qpack.HeaderField{
		{Name: ":method", Value: "GET"},
		{Name: ":scheme", Value: "https"},
		{Name: ":authority", Value: "user@example.com"},
		{Name: ":path", Value: "/"},
	}
	decode := func() (qpack.HeaderField, error) {
		if len(fields) == 0 {
			return qpack.HeaderField{}, io.EOF
		}
		field := fields[0]
		fields = fields[1:]
		return field, nil
	}

	_, err := parseHeaders(decode, true, 1<<20, nil, dump.Dumpers{})
	if err == nil || err.Error() != "userinfo is not allowed in :authority" {
		t.Fatalf("got error %v, want userinfo rejection", err)
	}
}

func TestExtractAnnouncedTrailers(t *testing.T) {
	tests := []struct {
		name     string
		header   http.Header
		expected http.Header
	}{
		{
			name:     "no trailer header",
			header:   http.Header{"Content-Type": []string{"text/html"}},
			expected: nil,
		},
		{
			name:     "single trailer",
			header:   http.Header{"Trailer": []string{"X-Custom-1"}},
			expected: http.Header{"X-Custom-1": nil},
		},
		{
			name:     "multiple trailers comma-separated",
			header:   http.Header{"Trailer": []string{"X-Custom-1, X-Custom-2"}},
			expected: http.Header{"X-Custom-1": nil, "X-Custom-2": nil},
		},
		{
			name:     "multiple trailers duplicate headers",
			header:   http.Header{"Trailer": []string{"X-Custom-1"}, "Trailer2": []string{"X-Custom-2"}},
			expected: http.Header{"X-Custom-1": nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAnnouncedTrailers(tt.header)
			if tt.expected == nil {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Fatalf("expected %v, got nil", tt.expected)
			}
			for k := range tt.expected {
				if _, ok := result[k]; !ok {
					t.Fatalf("expected key %s in result", k)
				}
			}
			// Verify "Trailer" header was removed
			if _, ok := tt.header["Trailer"]; ok {
				t.Fatal("Trailer header should have been removed")
			}
		})
	}
}

func TestExtractAnnouncedTrailersRemovesTrailerHeader(t *testing.T) {
	h := http.Header{
		"Content-Type": []string{"text/html"},
		"Trailer":      []string{"X-Custom-Trailer"},
	}
	result := extractAnnouncedTrailers(h)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result["X-Custom-Trailer"]; !ok {
		t.Fatal("expected X-Custom-Trailer in result")
	}
	if _, ok := h["Trailer"]; ok {
		t.Fatal("Trailer header should have been removed from original header")
	}
	if _, ok := h["Content-Type"]; !ok {
		t.Fatal("Content-Type should still be present")
	}
}
