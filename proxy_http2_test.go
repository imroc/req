package req

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// countingConnectProxy is a minimal HTTP CONNECT proxy for tests, which
// counts CONNECT requests and relays tunnel bytes.
type countingConnectProxy struct {
	ln           net.Listener
	connectCount int32

	mu      sync.Mutex
	tunnels []net.Conn
}

// startCountingConnectProxy starts a CONNECT proxy which relays tunnels to
// their requested targets, and registers cleanup to shut it down.
func startCountingConnectProxy(t *testing.T) *countingConnectProxy {
	t.Helper()
	p := &countingConnectProxy{}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p.ln = ln
	t.Cleanup(func() { ln.Close() })
	go p.serve()
	return p
}

// serve accepts and handles proxy connections until the listener is closed.
func (p *countingConnectProxy) serve() {
	for {
		cli, err := p.ln.Accept()
		if err != nil {
			return
		}
		go p.handle(cli)
	}
}

// handle establishes a tunnel for one CONNECT request and relays bytes.
func (p *countingConnectProxy) handle(cli net.Conn) {
	defer cli.Close()
	req, err := http.ReadRequest(bufio.NewReader(cli))
	if err != nil || req.Method != http.MethodConnect {
		return
	}
	atomic.AddInt32(&p.connectCount, 1)
	up, err := net.Dial("tcp", req.Host)
	if err != nil {
		cli.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer up.Close()
	p.mu.Lock()
	p.tunnels = append(p.tunnels, up)
	p.mu.Unlock()
	cli.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	go io.Copy(up, cli)
	io.Copy(cli, up)
}

// proxyURL returns the HTTP proxy URL of the proxy.
func (p *countingConnectProxy) proxyURL() string {
	return "http://" + p.ln.Addr().String()
}

// Connects returns how many CONNECT requests the proxy has received.
func (p *countingConnectProxy) Connects() int32 {
	return atomic.LoadInt32(&p.connectCount)
}

// killTunnels closes all established tunnels to the targets.
func (p *countingConnectProxy) killTunnels() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.tunnels {
		c.Close()
	}
	p.tunnels = nil
}

// TestHTTP2ProxyNoDirectDialOnTunnelDeath verifies that when an HTTP/2
// connection established through an HTTP proxy becomes unusable, subsequent
// requests dial a new connection through the proxy again instead of
// bypassing the proxy with a direct connection to the origin.
func TestHTTP2ProxyNoDirectDialOnTunnelDeath(t *testing.T) {
	// Backend HTTPS server with HTTP/2 enabled.
	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello-h2-proxy"))
	}))
	backend.EnableHTTP2 = true
	backend.StartTLS()
	defer backend.Close()

	proxy := startCountingConnectProxy(t)

	client := C().
		SetProxyURL(proxy.proxyURL()).
		EnableInsecureSkipVerify()

	// First request establishes the HTTP/2 connection through the proxy.
	resp, err := client.R().Get(backend.URL)
	if err != nil {
		t.Fatalf("first request via proxy failed: %v", err)
	}
	if resp.ProtoMajor != 2 {
		t.Fatalf("first request proto = %s; want HTTP/2", resp.Proto)
	}
	if n := proxy.Connects(); n != 1 {
		t.Fatalf("CONNECT count after first request = %d; want 1", n)
	}

	// Kill the tunnel, making the pooled HTTP/2 connection unusable.
	proxy.killTunnels()
	time.Sleep(300 * time.Millisecond)

	// The next request must dial a new tunnel through the proxy (CONNECT #2),
	// never a direct connection to the backend.
	resp, err = client.R().Get(backend.URL)
	if err != nil {
		t.Fatalf("second request via proxy failed: %v", err)
	}
	if resp.String() != "hello-h2-proxy" {
		t.Fatalf("body = %q; want hello-h2-proxy", resp.String())
	}
	if n := proxy.Connects(); n != 2 {
		t.Fatalf("CONNECT count after tunnel death = %d; want 2 (proxy was bypassed)", n)
	}
}

// TestForceHTTP2ViaProxy verifies that EnableForceHTTP2 honors the
// configured proxy instead of dialing the origin directly.
func TestForceHTTP2ViaProxy(t *testing.T) {
	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello-force-h2"))
	}))
	backend.EnableHTTP2 = true
	backend.StartTLS()
	defer backend.Close()

	proxy := startCountingConnectProxy(t)

	client := C().
		SetProxyURL(proxy.proxyURL()).
		EnableInsecureSkipVerify().
		EnableForceHTTP2()

	for i := 0; i < 2; i++ {
		resp, err := client.R().Get(backend.URL)
		if err != nil {
			t.Fatalf("request #%d via proxy failed: %v", i+1, err)
		}
		if resp.ProtoMajor != 2 {
			t.Fatalf("request #%d proto = %s; want HTTP/2", i+1, resp.Proto)
		}
	}
	if n := proxy.Connects(); n != 1 {
		t.Fatalf("CONNECT count = %d; want 1 (forced HTTP/2 must go through the proxy)", n)
	}
}

// TestForceHTTP2ViaProxyRejectsHTTP1Server verifies that with
// EnableForceHTTP2 and a proxy configured, a server which does not support
// HTTP/2 fails the request after dialing through the proxy, not by
// bypassing it.
func TestForceHTTP2ViaProxyRejectsHTTP1Server(t *testing.T) {
	// HTTP/1.1 only TLS server.
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello-h1"))
	}))
	defer backend.Close()

	proxy := startCountingConnectProxy(t)

	client := C().
		SetProxyURL(proxy.proxyURL()).
		EnableInsecureSkipVerify().
		EnableForceHTTP2()

	_, err := client.R().Get(backend.URL)
	if err == nil || !strings.Contains(err.Error(), "server does not support http2") {
		t.Fatalf("err = %v; want an error containing %q", err, "server does not support http2")
	}
	if n := proxy.Connects(); n != 1 {
		t.Fatalf("CONNECT count = %d; want 1 (proxy was bypassed)", n)
	}
}
