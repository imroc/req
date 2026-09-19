package req

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestForceHTTP3WithProxyFails verifies that EnableForceHTTP3 with a proxy
// configured fails explicitly instead of silently bypassing the proxy,
// since HTTP/3 (QUIC over UDP) cannot traverse a proxy.
func TestForceHTTP3WithProxyFails(t *testing.T) {
	proxy := startCountingConnectProxy(t)

	client := C().
		SetProxyURL(proxy.proxyURL()).
		EnableInsecureSkipVerify().
		EnableForceHTTP3()

	_, err := client.R().Get("https://127.0.0.1:1/")
	if err == nil || !strings.Contains(err.Error(), "cannot be used with a proxy") {
		t.Fatalf("err = %v; want an error containing %q", err, "cannot be used with a proxy")
	}
	if n := proxy.Connects(); n != 0 {
		t.Fatalf("CONNECT count = %d; want 0", n)
	}
}

// TestHTTP3AltSvcSkippedWithProxy verifies that with a proxy configured, an
// Alt-Svc advertisement for HTTP/3 is ignored: no direct QUIC connection to
// the origin is attempted, since HTTP/3 cannot traverse a proxy.
func TestHTTP3AltSvcSkippedWithProxy(t *testing.T) {
	// Black-hole UDP listener counting incoming packets (QUIC probes).
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer udpConn.Close()
	var udpPackets int32
	go func() {
		buf := make([]byte, 2048)
		for {
			if _, _, err := udpConn.ReadFromUDP(buf); err != nil {
				return
			}
			atomic.AddInt32(&udpPackets, 1)
		}
	}()

	// Backend HTTPS server advertising HTTP/3 at the black-hole UDP address.
	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Alt-Svc", fmt.Sprintf(`h3="%s"`, udpConn.LocalAddr().String()))
		w.Write([]byte("hello-h3-altsvc"))
	}))
	backend.EnableHTTP2 = true
	backend.StartTLS()
	defer backend.Close()

	proxy := startCountingConnectProxy(t)

	client := C().
		SetProxyURL(proxy.proxyURL()).
		EnableInsecureSkipVerify().
		EnableHTTP3()

	resp, err := client.R().Get(backend.URL)
	if err != nil {
		t.Fatalf("request via proxy failed: %v", err)
	}
	if resp.String() != "hello-h3-altsvc" {
		t.Fatalf("body = %q; want hello-h3-altsvc", resp.String())
	}

	// Give the alt-svc upgrade a chance to (wrongly) probe the QUIC address.
	time.Sleep(time.Second)
	if n := atomic.LoadInt32(&udpPackets); n != 0 {
		t.Fatalf("received %d UDP packets on the alt-svc address; want 0 (proxy must disable HTTP/3)", n)
	}
}
