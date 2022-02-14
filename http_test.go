package req

import "testing"

func TestCleanHost(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"www.google.com", "www.google.com"},
		{"www.google.com foo", "www.google.com"},
		{"www.google.com/foo", "www.google.com"},
		{" first character is a space", ""},
		{"[1::6]:8080", "[1::6]:8080"},

		// Punycode:
		{"гофер.рф/foo", "xn--c1ae0ajs.xn--p1ai"},
		{"bücher.de", "xn--bcher-kva.de"},
		{"bücher.de:8080", "xn--bcher-kva.de:8080"},
		// Verify we convert to lowercase before punycode:
		{"BÜCHER.de", "xn--bcher-kva.de"},
		{"BÜCHER.de:8080", "xn--bcher-kva.de:8080"},
		// Verify we normalize to NFC before punycode:
		{"gophér.nfc", "xn--gophr-esa.nfc"},            // NFC input; no work needed
		{"goph\u0065\u0301r.nfd", "xn--gophr-esa.nfd"}, // NFD input
	}
	for _, tt := range tests {
		got := cleanHost(tt.in)
		if tt.want != got {
			t.Errorf("cleanHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
