// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socks

import (
	"context"
	"errors"
	"golang.org/x/net/nettest"
	"io"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"
)

// An AuthRequest represents an authentication request.
type AuthRequest struct {
	Version int
	Methods []AuthMethod
}

// ParseAuthRequest parses an authentication request.
func ParseAuthRequest(b []byte) (*AuthRequest, error) {
	if len(b) < 2 {
		return nil, errors.New("short auth request")
	}
	if b[0] != Version5 {
		return nil, errors.New("unexpected protocol version")
	}
	if len(b)-2 < int(b[1]) {
		return nil, errors.New("short auth request")
	}
	req := &AuthRequest{Version: int(b[0])}
	if b[1] > 0 {
		req.Methods = make([]AuthMethod, b[1])
		for i, m := range b[2 : 2+b[1]] {
			req.Methods[i] = AuthMethod(m)
		}
	}
	return req, nil
}

// MarshalAuthReply returns an authentication reply in wire format.
func MarshalAuthReply(ver int, m AuthMethod) ([]byte, error) {
	return []byte{byte(ver), byte(m)}, nil
}

// A CmdRequest repesents a command request.
type CmdRequest struct {
	Version int
	Cmd     Command
	Addr    Addr
}

// ParseCmdRequest parses a command request.
func ParseCmdRequest(b []byte) (*CmdRequest, error) {
	if len(b) < 7 {
		return nil, errors.New("short cmd request")
	}
	if b[0] != Version5 {
		return nil, errors.New("unexpected protocol version")
	}
	if Command(b[1]) != CmdConnect {
		return nil, errors.New("unexpected command")
	}
	if b[2] != 0 {
		return nil, errors.New("non-zero reserved field")
	}
	req := &CmdRequest{Version: int(b[0]), Cmd: Command(b[1])}
	l := 2
	off := 4
	switch b[3] {
	case AddrTypeIPv4:
		l += net.IPv4len
		req.Addr.IP = make(net.IP, net.IPv4len)
	case AddrTypeIPv6:
		l += net.IPv6len
		req.Addr.IP = make(net.IP, net.IPv6len)
	case AddrTypeFQDN:
		l += int(b[4])
		off = 5
	default:
		return nil, errors.New("unknown address type")
	}
	if len(b[off:]) < l {
		return nil, errors.New("short cmd request")
	}
	if req.Addr.IP != nil {
		copy(req.Addr.IP, b[off:])
	} else {
		req.Addr.Name = string(b[off : off+l-2])
	}
	req.Addr.Port = int(b[off+l-2])<<8 | int(b[off+l-1])
	return req, nil
}

// MarshalCmdReply returns a command reply in wire format.
func MarshalCmdReply(ver int, reply Reply, a *Addr) ([]byte, error) {
	b := make([]byte, 4)
	b[0] = byte(ver)
	b[1] = byte(reply)
	if a.Name != "" {
		if len(a.Name) > 255 {
			return nil, errors.New("fqdn too long")
		}
		b[3] = AddrTypeFQDN
		b = append(b, byte(len(a.Name)))
		b = append(b, a.Name...)
	} else if ip4 := a.IP.To4(); ip4 != nil {
		b[3] = AddrTypeIPv4
		b = append(b, ip4...)
	} else if ip6 := a.IP.To16(); ip6 != nil {
		b[3] = AddrTypeIPv6
		b = append(b, ip6...)
	} else {
		return nil, errors.New("unknown address type")
	}
	b = append(b, byte(a.Port>>8), byte(a.Port))
	return b, nil
}

// A Server repesents a server for handshake testing.
type Server struct {
	ln net.Listener
}

// Addr rerurns a server address.
func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}

// TargetAddr returns a fake final destination address.
//
// The returned address is only valid for testing with Server.
func (s *Server) TargetAddr() net.Addr {
	a := s.ln.Addr()
	switch a := a.(type) {
	case *net.TCPAddr:
		if a.IP.To4() != nil {
			return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5963}
		}
		if a.IP.To16() != nil && a.IP.To4() == nil {
			return &net.TCPAddr{IP: net.IPv6loopback, Port: 5963}
		}
	}
	return nil
}

// Close closes the server.
func (s *Server) Close() error {
	return s.ln.Close()
}

func (s *Server) serve(authFunc, cmdFunc func(io.ReadWriter, []byte) error) {
	c, err := s.ln.Accept()
	if err != nil {
		return
	}
	defer c.Close()
	go s.serve(authFunc, cmdFunc)
	b := make([]byte, 512)
	n, err := c.Read(b)
	if err != nil {
		return
	}
	if err := authFunc(c, b[:n]); err != nil {
		return
	}
	n, err = c.Read(b)
	if err != nil {
		return
	}
	if err := cmdFunc(c, b[:n]); err != nil {
		return
	}
}

// NewServer returns a new server.
//
// The provided authFunc and cmdFunc must parse requests and return
// appropriate replies to clients.
func NewServer(authFunc, cmdFunc func(io.ReadWriter, []byte) error) (*Server, error) {
	var err error
	s := new(Server)
	s.ln, err = nettest.NewLocalListener("tcp")
	if err != nil {
		return nil, err
	}
	go s.serve(authFunc, cmdFunc)
	return s, nil
}

// NoAuthRequired handles a no-authentication-required signaling.
func NoAuthRequired(rw io.ReadWriter, b []byte) error {
	req, err := ParseAuthRequest(b)
	if err != nil {
		return err
	}
	b, err = MarshalAuthReply(req.Version, AuthMethodNotRequired)
	if err != nil {
		return err
	}
	n, err := rw.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errors.New("short write")
	}
	return nil
}

// NoProxyRequired handles a command signaling without constructing a
// proxy connection to the final destination.
func NoProxyRequired(rw io.ReadWriter, b []byte) error {
	req, err := ParseCmdRequest(b)
	if err != nil {
		return err
	}
	req.Addr.Port += 1
	if req.Addr.Name != "" {
		req.Addr.Name = "boundaddr.doesnotexist"
	} else if req.Addr.IP.To4() != nil {
		req.Addr.IP = net.IPv4(127, 0, 0, 1)
	} else {
		req.Addr.IP = net.IPv6loopback
	}
	b, err = MarshalCmdReply(Version5, StatusSucceeded, &req.Addr)
	if err != nil {
		return err
	}
	n, err := rw.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errors.New("short write")
	}
	return nil
}

func TestDial(t *testing.T) {
	t.Run("Connect", func(t *testing.T) {
		ss, err := NewServer(NoAuthRequired, NoProxyRequired)
		if err != nil {
			t.Fatal(err)
		}
		defer ss.Close()
		d := NewDialer(ss.Addr().Network(), ss.Addr().String())
		d.AuthMethods = []AuthMethod{
			AuthMethodNotRequired,
			AuthMethodUsernamePassword,
		}
		d.Authenticate = (&UsernamePassword{
			Username: "username",
			Password: "password",
		}).Authenticate
		c, err := d.DialContext(context.Background(), ss.TargetAddr().Network(), ss.TargetAddr().String())
		if err != nil {
			t.Fatal(err)
		}
		c.(*Conn).BoundAddr()
		c.Close()
	})
	t.Run("ConnectWithConn", func(t *testing.T) {
		ss, err := NewServer(NoAuthRequired, NoProxyRequired)
		if err != nil {
			t.Fatal(err)
		}
		defer ss.Close()
		c, err := net.Dial(ss.Addr().Network(), ss.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		d := NewDialer(ss.Addr().Network(), ss.Addr().String())
		d.AuthMethods = []AuthMethod{
			AuthMethodNotRequired,
			AuthMethodUsernamePassword,
		}
		d.Authenticate = (&UsernamePassword{
			Username: "username",
			Password: "password",
		}).Authenticate
		a, err := d.DialWithConn(context.Background(), c, ss.TargetAddr().Network(), ss.TargetAddr().String())
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := a.(*Addr); !ok {
			t.Fatalf("got %+v; want Addr", a)
		}
	})
	t.Run("Cancel", func(t *testing.T) {
		ss, err := NewServer(NoAuthRequired, blackholeCmdFunc)
		if err != nil {
			t.Fatal(err)
		}
		defer ss.Close()
		d := NewDialer(ss.Addr().Network(), ss.Addr().String())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dialErr := make(chan error)
		go func() {
			c, err := d.DialContext(ctx, ss.TargetAddr().Network(), ss.TargetAddr().String())
			if err == nil {
				c.Close()
			}
			dialErr <- err
		}()
		time.Sleep(100 * time.Millisecond)
		cancel()
		err = <-dialErr
		if perr, nerr := parseDialError(err); perr != context.Canceled && nerr == nil {
			t.Fatalf("got %v; want context.Canceled or equivalent", err)
		}
	})
	t.Run("Deadline", func(t *testing.T) {
		ss, err := NewServer(NoAuthRequired, blackholeCmdFunc)
		if err != nil {
			t.Fatal(err)
		}
		defer ss.Close()
		d := NewDialer(ss.Addr().Network(), ss.Addr().String())
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(100*time.Millisecond))
		defer cancel()
		c, err := d.DialContext(ctx, ss.TargetAddr().Network(), ss.TargetAddr().String())
		if err == nil {
			c.Close()
		}
		if perr, nerr := parseDialError(err); perr != context.DeadlineExceeded && nerr == nil {
			t.Fatalf("got %v; want context.DeadlineExceeded or equivalent", err)
		}
	})
	t.Run("WithRogueServer", func(t *testing.T) {
		ss, err := NewServer(NoAuthRequired, rogueCmdFunc)
		if err != nil {
			t.Fatal(err)
		}
		defer ss.Close()
		d := NewDialer(ss.Addr().Network(), ss.Addr().String())
		for i := 0; i < 2*len(rogueCmdList); i++ {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(100*time.Millisecond))
			defer cancel()
			c, err := d.DialContext(ctx, ss.TargetAddr().Network(), ss.TargetAddr().String())
			if err == nil {
				t.Log(c.(*Conn).BoundAddr())
				c.Close()
				t.Error("should fail")
			}
		}
	})
}

func blackholeCmdFunc(rw io.ReadWriter, b []byte) error {
	if _, err := ParseCmdRequest(b); err != nil {
		return err
	}
	var bb [1]byte
	for {
		if _, err := rw.Read(bb[:]); err != nil {
			return err
		}
	}
}

func rogueCmdFunc(rw io.ReadWriter, b []byte) error {
	if _, err := ParseCmdRequest(b); err != nil {
		return err
	}
	rw.Write(rogueCmdList[rand.Intn(len(rogueCmdList))])
	return nil
}

var rogueCmdList = [][]byte{
	{0x05},
	{0x06, 0x00, 0x00, 0x01, 192, 0, 2, 1, 0x17, 0x4b},
	{0x05, 0x00, 0xff, 0x01, 192, 0, 2, 2, 0x17, 0x4b},
	{0x05, 0x00, 0x00, 0x01, 192, 0, 2, 3},
	{0x05, 0x00, 0x00, 0x03, 0x04, 'F', 'Q', 'D', 'N'},
}

func parseDialError(err error) (perr, nerr error) {
	if e, ok := err.(*net.OpError); ok {
		err = e.Err
		nerr = e
	}
	if e, ok := err.(*os.SyscallError); ok {
		err = e.Err
	}
	perr = err
	return
}
