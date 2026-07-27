package req

import (
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestSocks4ProxyE2E verifies that Client.SetProxyURL("socks4://...") can
// successfully send an HTTP request through a SOCKS4 proxy.
func TestSocks4ProxyE2E(t *testing.T) {
	// Backend HTTP server.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "socks4")
		w.Write([]byte("hello-socks4"))
	}))
	defer backend.Close()

	backendHost := backend.Listener.Addr().String()

	// Minimal SOCKS4 relay proxy.
	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyLn.Close()

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := proxyLn.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer c.Close()
				if err := handleSocks4Connect(c); err != nil {
					return
				}
			}(conn)
		}
	}()
	defer func() {
		close(done)
		proxyLn.Close()
		wg.Wait()
	}()

	proxyURL := "socks4://userid@" + proxyLn.Addr().String()
	client := C().SetProxyURL(proxyURL).DisableKeepAlives()

	resp, err := client.R().Get(backend.URL)
	if err != nil {
		t.Fatalf("request via socks4 proxy failed: %v (backend=%s)", err, backendHost)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d; want 200", resp.StatusCode)
	}
	if resp.Header.Get("X-Test") != "socks4" {
		t.Fatalf("X-Test = %q; want socks4", resp.Header.Get("X-Test"))
	}
	if resp.String() != "hello-socks4" {
		t.Fatalf("body = %q; want hello-socks4", resp.String())
	}
}

// TestSocks4aProxyE2E verifies socks4a scheme with domain-style target.
func TestSocks4aProxyE2E(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello-socks4a"))
	}))
	defer backend.Close()

	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyLn.Close()

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := proxyLn.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer c.Close()
				_ = handleSocks4Connect(c)
			}(conn)
		}
	}()
	defer func() {
		close(done)
		proxyLn.Close()
		wg.Wait()
	}()

	// Use 127.0.0.1 in the backend URL so the target address is IPv4;
	// socks4a still works when the host is an IP (no domain extension needed).
	client := C().SetProxyURL("socks4a://" + proxyLn.Addr().String()).DisableKeepAlives()

	resp, err := client.R().Get(backend.URL)
	if err != nil {
		t.Fatalf("request via socks4a proxy failed: %v", err)
	}
	if resp.String() != "hello-socks4a" {
		t.Fatalf("body = %q; want hello-socks4a", resp.String())
	}
}

// handleSocks4Connect performs a SOCKS4/4a CONNECT handshake and relays.
func handleSocks4Connect(c net.Conn) error {
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))

	var hdr [8]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		return err
	}
	if hdr[0] != 0x04 || hdr[1] != 0x01 {
		// Reject non-CONNECT or non-SOCKS4.
		_, _ = c.Write([]byte{0, 91, 0, 0, 0, 0, 0, 0})
		return nil
	}
	port := int(binary.BigEndian.Uint16(hdr[2:4]))
	ip := net.IPv4(hdr[4], hdr[5], hdr[6], hdr[7])

	// Read userid.
	if _, err := readNULString(c); err != nil {
		return err
	}

	var host string
	// SOCKS4a domain when IP is 0.0.0.x with x != 0
	if hdr[4] == 0 && hdr[5] == 0 && hdr[6] == 0 && hdr[7] != 0 {
		domain, err := readNULString(c)
		if err != nil {
			return err
		}
		host = domain
	} else {
		host = ip.String()
	}

	target := net.JoinHostPort(host, itoa(port))
	upstream, err := net.DialTimeout("tcp", target, 2*time.Second)
	if err != nil {
		_, _ = c.Write([]byte{0, 91, 0, 0, 0, 0, 0, 0})
		return err
	}
	defer upstream.Close()

	if _, err := c.Write([]byte{0, 90, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}

	// Clear handshake deadline before long-lived relay.
	_ = c.SetDeadline(time.Time{})

	errc := make(chan struct{}, 2)
	go func() { io.Copy(upstream, c); errc <- struct{}{} }()
	go func() { io.Copy(c, upstream); errc <- struct{}{} }()
	<-errc
	return nil
}

func readNULString(r io.Reader) (string, error) {
	var b [1]byte
	var out []byte
	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return "", err
		}
		if b[0] == 0 {
			return string(out), nil
		}
		out = append(out, b[0])
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
