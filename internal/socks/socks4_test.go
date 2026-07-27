package socks

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// startSocks4Server starts a minimal SOCKS4/4a test server.
// handle is called with the parsed request fields and should return
// the reply code and optional destination to relay to. If dest is
// non-empty and the reply is Status4Granted, the server dials dest
// and bidirectionally copies data.
func startSocks4Server(t *testing.T, handle func(req socks4Request) (Reply, string)) (addr string, closeFn func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
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
				req, err := readSocks4Request(c)
				if err != nil {
					return
				}
				code, dest := handle(req)
				reply := make([]byte, 8)
				reply[0] = 0
				reply[1] = byte(code)
				if _, err := c.Write(reply); err != nil {
					return
				}
				if code != Status4Granted || dest == "" {
					return
				}
				upstream, err := net.DialTimeout("tcp", dest, 2*time.Second)
				if err != nil {
					return
				}
				defer upstream.Close()
				relay(c, upstream)
			}(conn)
		}
	}()

	return ln.Addr().String(), func() {
		close(done)
		ln.Close()
		wg.Wait()
	}
}

type socks4Request struct {
	Cmd    Command
	Port   int
	IP     net.IP
	UserID string
	Domain string // non-empty when SOCKS4a domain was provided
}

func readSocks4Request(r io.Reader) (socks4Request, error) {
	var hdr [8]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return socks4Request{}, err
	}
	if hdr[0] != Version4 {
		return socks4Request{}, io.ErrUnexpectedEOF
	}
	req := socks4Request{
		Cmd:  Command(hdr[1]),
		Port: int(binary.BigEndian.Uint16(hdr[2:4])),
		IP:   net.IPv4(hdr[4], hdr[5], hdr[6], hdr[7]).To4(),
	}

	userID, err := readCString(r)
	if err != nil {
		return socks4Request{}, err
	}
	req.UserID = userID

	// SOCKS4a: IP is 0.0.0.x with x != 0
	if req.IP[0] == 0 && req.IP[1] == 0 && req.IP[2] == 0 && req.IP[3] != 0 {
		domain, err := readCString(r)
		if err != nil {
			return socks4Request{}, err
		}
		req.Domain = domain
	}
	return req, nil
}

func readCString(r io.Reader) (string, error) {
	var buf bytes.Buffer
	var b [1]byte
	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return "", err
		}
		if b[0] == 0 {
			return buf.String(), nil
		}
		buf.WriteByte(b[0])
	}
}

func relay(a, b net.Conn) {
	done := make(chan struct{}, 2)
	copyFn := func(dst, src net.Conn) {
		io.Copy(dst, src)
		done <- struct{}{}
	}
	go copyFn(a, b)
	go copyFn(b, a)
	<-done
}

func TestSocks4ConnectIPv4(t *testing.T) {
	const userID = "tester"
	got := make(chan socks4Request, 1)

	addr, closeFn := startSocks4Server(t, func(req socks4Request) (Reply, string) {
		got <- req
		return Status4Granted, ""
	})
	defer closeFn()

	d := NewDialer("tcp", addr)
	d.Version = Version4
	d.UserID = userID

	c, err := d.DialContext(context.Background(), "tcp", "1.2.3.4:80")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	req := <-got
	if req.Cmd != CmdConnect {
		t.Fatalf("cmd = %v; want CONNECT", req.Cmd)
	}
	if req.Port != 80 {
		t.Fatalf("port = %d; want 80", req.Port)
	}
	if !req.IP.Equal(net.IPv4(1, 2, 3, 4)) {
		t.Fatalf("ip = %v; want 1.2.3.4", req.IP)
	}
	if req.UserID != userID {
		t.Fatalf("userid = %q; want %q", req.UserID, userID)
	}
	if req.Domain != "" {
		t.Fatalf("domain = %q; want empty", req.Domain)
	}
}

func TestSocks4aConnectDomain(t *testing.T) {
	got := make(chan socks4Request, 1)

	addr, closeFn := startSocks4Server(t, func(req socks4Request) (Reply, string) {
		got <- req
		return Status4Granted, ""
	})
	defer closeFn()

	d := NewDialer("tcp", addr)
	d.Version = Version4
	d.Socks4A = true
	d.UserID = "alice"

	c, err := d.DialContext(context.Background(), "tcp", "example.com:443")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	req := <-got
	if req.Domain != "example.com" {
		t.Fatalf("domain = %q; want example.com", req.Domain)
	}
	if req.Port != 443 {
		t.Fatalf("port = %d; want 443", req.Port)
	}
	if req.UserID != "alice" {
		t.Fatalf("userid = %q; want alice", req.UserID)
	}
	// SOCKS4a uses 0.0.0.x with x != 0
	if req.IP[0] != 0 || req.IP[1] != 0 || req.IP[2] != 0 || req.IP[3] == 0 {
		t.Fatalf("ip = %v; want 0.0.0.x with x != 0", req.IP)
	}
}

func TestSocks4EmptyUserID(t *testing.T) {
	got := make(chan socks4Request, 1)
	addr, closeFn := startSocks4Server(t, func(req socks4Request) (Reply, string) {
		got <- req
		return Status4Granted, ""
	})
	defer closeFn()

	d := NewDialer("tcp", addr)
	d.Version = Version4

	c, err := d.DialContext(context.Background(), "tcp", "127.0.0.1:9")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	req := <-got
	if req.UserID != "" {
		t.Fatalf("userid = %q; want empty", req.UserID)
	}
}

func TestSocks4Rejected(t *testing.T) {
	addr, closeFn := startSocks4Server(t, func(req socks4Request) (Reply, string) {
		return Status4Rejected, ""
	})
	defer closeFn()

	d := NewDialer("tcp", addr)
	d.Version = Version4

	_, err := d.DialContext(context.Background(), "tcp", "127.0.0.1:9")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), Status4Rejected.String()) {
		t.Fatalf("error = %v; want containing %q", err, Status4Rejected.String())
	}
}

func TestSocks4IPv6Rejected(t *testing.T) {
	// Server should not be needed; client rejects IPv6 before writing.
	d := NewDialer("tcp", "127.0.0.1:1")
	d.Version = Version4

	// DialWithConn uses an already-open connection, so we can test the
	// handshake without a real proxy listener.
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	errCh := make(chan error, 1)
	go func() {
		_, err := d.DialWithConn(context.Background(), client, "tcp", "[::1]:80")
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error for IPv6")
		}
		if !strings.Contains(err.Error(), "IPv6") {
			t.Fatalf("error = %v; want IPv6 mention", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestSocks4InvalidUserID(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	d := NewDialer("tcp", "127.0.0.1:1")
	d.Version = Version4
	d.UserID = "bad\x00id"

	_, err := d.DialWithConn(context.Background(), client, "tcp", "127.0.0.1:80")
	if err == nil {
		t.Fatal("expected error for NUL in user ID")
	}
	if !strings.Contains(err.Error(), "NUL") {
		t.Fatalf("error = %v; want NUL mention", err)
	}
}

func TestSocks4aInvalidDomain(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	d := NewDialer("tcp", "127.0.0.1:1")
	d.Version = Version4
	d.Socks4A = true

	_, err := d.DialWithConn(context.Background(), client, "tcp", "bad\x00host:80")
	if err == nil {
		t.Fatal("expected error for NUL in domain")
	}
	if !strings.Contains(err.Error(), "NUL") {
		t.Fatalf("error = %v; want NUL mention", err)
	}
}

func TestSocks4aDomainTooLong(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	d := NewDialer("tcp", "127.0.0.1:1")
	d.Version = Version4
	d.Socks4A = true

	longHost := strings.Repeat("a", maxSocks4aDomainLen+1) + ":80"
	_, err := d.DialWithConn(context.Background(), client, "tcp", longHost)
	if err == nil {
		t.Fatal("expected error for long domain")
	}
	if !strings.Contains(err.Error(), "too long") {
		t.Fatalf("error = %v; want too long", err)
	}
}

func TestSocks4LocalResolve(t *testing.T) {
	got := make(chan socks4Request, 1)
	addr, closeFn := startSocks4Server(t, func(req socks4Request) (Reply, string) {
		got <- req
		return Status4Granted, ""
	})
	defer closeFn()

	d := NewDialer("tcp", addr)
	d.Version = Version4
	// Socks4A is false: domain names must be resolved locally.

	c, err := d.DialContext(context.Background(), "tcp", "localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	req := <-got
	if req.Domain != "" {
		t.Fatalf("domain = %q; want empty (local resolve)", req.Domain)
	}
	if req.Port != 8080 {
		t.Fatalf("port = %d; want 8080", req.Port)
	}
	if !req.IP.IsLoopback() {
		t.Fatalf("ip = %v; want loopback", req.IP)
	}
}

func TestSocks4DialWithConn(t *testing.T) {
	got := make(chan socks4Request, 1)
	addr, closeFn := startSocks4Server(t, func(req socks4Request) (Reply, string) {
		got <- req
		return Status4Granted, ""
	})
	defer closeFn()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	d := NewDialer("tcp", addr)
	d.Version = Version4
	d.Socks4A = true
	d.UserID = "withconn"

	a, err := d.DialWithConn(context.Background(), conn, "tcp", "service.local:1234")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.(*Addr); !ok {
		t.Fatalf("got %T; want *Addr", a)
	}

	req := <-got
	if req.Domain != "service.local" {
		t.Fatalf("domain = %q; want service.local", req.Domain)
	}
	if req.Port != 1234 {
		t.Fatalf("port = %d; want 1234", req.Port)
	}
	if req.UserID != "withconn" {
		t.Fatalf("userid = %q; want withconn", req.UserID)
	}
}

func TestSocks4ReplyVersionTolerance(t *testing.T) {
	// Some servers incorrectly set VN=4 in the reply; client should accept it.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		// Read full request (header + userid NUL)
		buf := make([]byte, 256)
		n, _ := c.Read(buf)
		_ = n
		// Reply with VN=4 instead of 0
		c.Write([]byte{Version4, byte(Status4Granted), 0, 0, 0, 0, 0, 0})
	}()

	d := NewDialer("tcp", ln.Addr().String())
	d.Version = Version4
	c, err := d.DialContext(context.Background(), "tcp", "10.0.0.1:80")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
}

func TestSocks4UnexpectedVersion(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		buf := make([]byte, 256)
		c.Read(buf)
		c.Write([]byte{5, byte(Status4Granted), 0, 0, 0, 0, 0, 0})
	}()

	d := NewDialer("tcp", ln.Addr().String())
	d.Version = Version4
	_, err = d.DialContext(context.Background(), "tcp", "10.0.0.1:80")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unexpected protocol version") {
		t.Fatalf("error = %v; want unexpected protocol version", err)
	}
}

func TestReply4String(t *testing.T) {
	if got, want := Status4Granted.String(), "request granted"; got != want {
		t.Errorf("Status4Granted = %q; want %q", got, want)
	}
	if got, want := Status4Rejected.String(), "request rejected or failed"; got != want {
		t.Errorf("Status4Rejected = %q; want %q", got, want)
	}
	if s := Status4IdentdFailed.String(); !strings.Contains(s, "identd") {
		t.Errorf("Status4IdentdFailed = %q; want containing identd", s)
	}
	if s := Status4IdentdMismatch.String(); !strings.Contains(s, "user-ids") {
		t.Errorf("Status4IdentdMismatch = %q; want containing user-ids", s)
	}
}

func TestSocks4RelayHTTP(t *testing.T) {
	// End-to-end: SOCKS4 proxy relays to a real TCP echo/HTTP target.
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer targetLn.Close()

	const payload = "HTTP/1.0 200 OK\r\nContent-Length: 2\r\n\r\nOK"
	go func() {
		c, err := targetLn.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		// Read the request once, then respond. Do not wait for EOF/full buffer.
		buf := make([]byte, 256)
		_, _ = c.Read(buf)
		_, _ = c.Write([]byte(payload))
	}()

	targetAddr := targetLn.Addr().String()
	host, portStr, _ := net.SplitHostPort(targetAddr)
	port, _ := strconv.Atoi(portStr)

	proxyAddr, closeFn := startSocks4Server(t, func(req socks4Request) (Reply, string) {
		if req.Port != port {
			return Status4Rejected, ""
		}
		if !req.IP.Equal(net.ParseIP(host)) {
			return Status4Rejected, ""
		}
		return Status4Granted, targetAddr
	})
	defer closeFn()

	d := NewDialer("tcp", proxyAddr)
	d.Version = Version4

	c, err := d.DialContext(context.Background(), "tcp", targetAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if _, err := c.Write([]byte("GET / HTTP/1.0\r\n\r\n")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 128)
	n, err := c.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buf[:n]), "OK") {
		t.Fatalf("response = %q; want containing OK", buf[:n])
	}
}
