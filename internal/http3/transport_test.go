package http3

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/imroc/req/v3/internal/testcert"
	"github.com/quic-go/quic-go"
)

func TestTransportInit(t *testing.T) {
	tr := &Transport{}
	// Trigger init by calling RoundTrip with invalid request
	// This tests that init() doesn't panic
	_, err := tr.RoundTrip(&http.Request{})
	if err == nil {
		t.Fatal("expected error for nil URL")
	}
}

func TestTransportInitWithDatagrams(t *testing.T) {
	tr := &Transport{
		EnableDatagrams: true,
		QUICConfig: &quic.Config{
			EnableDatagrams: true,
		},
	}
	_, err := tr.RoundTrip(&http.Request{})
	if err == nil {
		t.Fatal("expected error for nil URL")
	}
}

func TestTransportInitDatagramMismatch(t *testing.T) {
	tr := &Transport{
		EnableDatagrams: true,
		QUICConfig: &quic.Config{
			EnableDatagrams: false,
		},
	}
	_, err := tr.RoundTrip(&http.Request{})
	if err == nil || err.Error() != "HTTP Datagrams enabled, but QUIC Datagrams disabled" {
		t.Fatalf("expected datagram mismatch error, got: %v", err)
	}
}

func TestRawClientConnType(t *testing.T) {
	// Test that RawClientConn embeds ClientConn correctly
	// This is a compile-time type check
	type hasClientConn interface {
		RoundTrip(*http.Request) (*http.Response, error)
	}
	var _ hasClientConn = (*RawClientConn)(nil)
	var _ hasClientConn = (*ClientConn)(nil)
}

func TestNewClientConnAndRawClientConn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create TLS config from test certificates
	cert, err := tls.X509KeyPair(testcert.LocalhostCert, testcert.LocalhostKey)
	if err != nil {
		t.Fatalf("failed to load test cert: %v", err)
	}
	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3"},
	}

	// Start QUIC listener
	listener, err := quic.ListenAddr("127.0.0.1:0", serverTLSConfig, &quic.Config{})
	if err != nil {
		t.Fatalf("failed to start QUIC listener: %v", err)
	}
	defer listener.Close()

	serverAddr := listener.Addr().String()

	// Accept connections in background
	go func() {
		for {
			conn, err := listener.Accept(context.Background())
			if err != nil {
				return
			}
			conn.CloseWithError(0, "test done")
		}
	}()

	// Create client transport
	clientTLSConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},
	}

	tr := &Transport{
		TLSClientConfig: clientTLSConfig,
		QUICConfig: &quic.Config{
			MaxIdleTimeout: 5 * time.Second,
		},
	}
	defer tr.Close()

	// Test dialing a QUIC connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quicConn, err := quic.DialAddr(ctx, serverAddr, clientTLSConfig, tr.QUICConfig)
	if err != nil {
		t.Fatalf("failed to dial QUIC: %v", err)
	}
	defer quicConn.CloseWithError(0, "")

	// Test NewClientConn
	clientConn := tr.NewClientConn(quicConn)
	if clientConn == nil {
		t.Fatal("NewClientConn returned nil")
	}

	// Test NewRawClientConn with a new connection
	quicConn2, err := quic.DialAddr(ctx, serverAddr, clientTLSConfig, tr.QUICConfig)
	if err != nil {
		t.Fatalf("failed to dial QUIC for raw conn: %v", err)
	}
	defer quicConn2.CloseWithError(0, "")

	rawConn := tr.NewRawClientConn(quicConn2)
	if rawConn == nil {
		t.Fatal("NewRawClientConn returned nil")
	}
	if rawConn.ClientConn == nil {
		t.Fatal("RawClientConn.ClientConn is nil")
	}

	t.Log("Successfully created ClientConn and RawClientConn")
}

// AddConn is the second way into getClient, beside roundTripOpt, and only
// roundTripOpt ran init. A transport whose first use is AddConn therefore
// dialed with a nil QUICConfig, a nil quic.Transport and a nil newClientConn,
// and the dial runs in its own goroutine, so the nil transport took the
// process down with it.
func TestAddConnInitializesTheTransport(t *testing.T) {
	dialed := make(chan *quic.Config, 1)
	tr := &Transport{
		Dial: func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
			dialed <- cfg
			return nil, errors.New("this test does not dial")
		},
	}

	if err := tr.AddConn(context.Background(), "127.0.0.1:443"); err != nil {
		t.Fatalf("AddConn: %v", err)
	}

	var cfg *quic.Config
	select {
	case cfg = <-dialed:
	case <-time.After(10 * time.Second):
		t.Fatal("the dial never ran")
	}

	if cfg == nil {
		t.Fatal("dialed with a nil quic.Config, so init did not run before AddConn dialed")
	}
	if cfg.MaxIncomingStreams != -1 {
		t.Errorf("MaxIncomingStreams is %d, want -1", cfg.MaxIncomingStreams)
	}
	if len(cfg.Versions) != 1 {
		t.Errorf("Versions is %v, want exactly one", cfg.Versions)
	}
}
