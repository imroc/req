package req

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/imroc/req/v3/internal/testcert"
	"github.com/quic-go/quic-go/http3"
)

// startHTTP3Server serves handler over HTTP/3 with the self-signed test
// certificate and returns its address.
func startHTTP3Server(t *testing.T, handler http.Handler) string {
	t.Helper()
	cert, err := tls.X509KeyPair(testcert.LocalhostCert, testcert.LocalhostKey)
	if err != nil {
		t.Fatal(err)
	}
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http3.Server{
		Handler:   handler,
		TLSConfig: http3.ConfigureTLSConfig(&tls.Config{Certificates: []tls.Certificate{cert}}),
	}
	go srv.Serve(pc)
	t.Cleanup(func() { srv.Close() })
	return "https://" + pc.LocalAddr().String() + "/"
}

// The HTTP/3 transport used to keep a TLSClientConfig of its own, which was
// never set and hid the client's, so InsecureSkipVerify, RootCAs and client
// certificates did nothing over HTTP/3.
func TestHTTP3UsesTheClientTLSConfig(t *testing.T) {
	url := startHTTP3Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, r.Proto)
	}))
	resp, err := C().EnableForceHTTP3().EnableInsecureSkipVerify().
		SetTimeout(5 * time.Second).R().Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if got := resp.String(); got != "HTTP/3.0" {
		t.Fatalf("body = %q, want HTTP/3.0", got)
	}
}
