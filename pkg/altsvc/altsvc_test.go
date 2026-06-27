package altsvc

import (
	"testing"
	"time"
)

func TestAltSvcJarSetGet(t *testing.T) {
	jar := NewAltSvcJar()
	as := &AltSvc{
		Protocol: "h3",
		Host:     "",
		Port:     "8443",
		Expire:   time.Now().Add(1 * time.Hour),
	}
	jar.SetAltSvc("example.com:443", as)

	got := jar.GetAltSvc("example.com:443")
	if got == nil {
		t.Fatal("expected non-nil AltSvc")
	}
	if got.Protocol != "h3" {
		t.Fatalf("Protocol = %q, want h3", got.Protocol)
	}
	if got.Port != "8443" {
		t.Fatalf("Port = %q, want 8443", got.Port)
	}
}

func TestAltSvcJarExpired(t *testing.T) {
	jar := NewAltSvcJar()
	as := &AltSvc{
		Protocol: "h3",
		Port:     "8443",
		Expire:   time.Now().Add(-1 * time.Hour), // expired
	}
	jar.SetAltSvc("example.com:443", as)

	got := jar.GetAltSvc("example.com:443")
	if got != nil {
		t.Fatal("expected nil for expired AltSvc")
	}
}

func TestAltSvcJarEmptyAddr(t *testing.T) {
	jar := NewAltSvcJar()
	jar.SetAltSvc("", &AltSvc{Protocol: "h3"})
	if got := jar.GetAltSvc(""); got != nil {
		t.Fatal("expected nil for empty addr")
	}
}

func TestAltSvcJarNotFound(t *testing.T) {
	jar := NewAltSvcJar()
	if got := jar.GetAltSvc("nonexistent.com:443"); got != nil {
		t.Fatal("expected nil for non-existent addr")
	}
}
