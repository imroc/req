package transport

import (
	"crypto/tls"
	"testing"
)

func TestOptionsClone(t *testing.T) {
	opt := Options{
		DisableKeepAlives:   true,
		MaxIdleConns:        10,
		MaxConnsPerHost:     5,
		IdleConnTimeout:     30000000000,
		TLSHandshakeTimeout: 10000000000,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "example.com",
		},
	}

	cloned := opt.Clone()
	if cloned.DisableKeepAlives != true {
		t.Fatal("DisableKeepAlives mismatch")
	}
	if cloned.MaxIdleConns != 10 {
		t.Fatal("MaxIdleConns mismatch")
	}
	if cloned.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil after clone")
	}
	if !cloned.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("TLSClientConfig.InsecureSkipVerify mismatch")
	}
	// Verify it's a different pointer
	if &cloned.TLSClientConfig == &opt.TLSClientConfig {
		t.Fatal("TLSClientConfig should be a new pointer after clone")
	}
}

func TestOptionsCloneNilConfig(t *testing.T) {
	opt := Options{
		MaxIdleConns: 10,
	}
	cloned := opt.Clone()
	if cloned.TLSClientConfig != nil {
		t.Fatal("TLSClientConfig should be nil")
	}
	if cloned.MaxIdleConns != 10 {
		t.Fatal("MaxIdleConns mismatch")
	}
}
