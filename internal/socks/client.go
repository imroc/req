// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socks

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"time"
)

var (
	noDeadline   = time.Time{}
	aLongTimeAgo = time.Unix(1, 0)
)

func (d *Dialer) connect(ctx context.Context, c net.Conn, address string) (_ net.Addr, ctxErr error) {
	host, port, err := splitHostPort(address)
	if err != nil {
		return nil, err
	}
	if deadline, ok := ctx.Deadline(); ok && !deadline.IsZero() {
		c.SetDeadline(deadline)
		defer c.SetDeadline(noDeadline)
	}
	if ctx != context.Background() {
		errCh := make(chan error, 1)
		done := make(chan struct{})
		defer func() {
			close(done)
			if ctxErr == nil {
				ctxErr = <-errCh
			}
		}()
		go func() {
			select {
			case <-ctx.Done():
				c.SetDeadline(aLongTimeAgo)
				errCh <- ctx.Err()
			case <-done:
				errCh <- nil
			}
		}()
	}

	var addr net.Addr
	if d.version() == Version4 {
		addr, ctxErr = d.connect4(ctx, c, host, port)
	} else {
		addr, ctxErr = d.connect5(ctx, c, host, port)
	}
	return addr, ctxErr
}

// maxSocks4aDomainLen is the maximum domain name length accepted for SOCKS4a.
// Matches the practical FQDN limit used by SOCKS5 in this package.
const maxSocks4aDomainLen = 255

// connect4 implements the SOCKS4 and SOCKS4a CONNECT handshake.
func (d *Dialer) connect4(ctx context.Context, c net.Conn, host string, port int) (net.Addr, error) {
	if err := validateSocks4CString(d.UserID, "user ID"); err != nil {
		return nil, err
	}

	var (
		ip     net.IP
		domain string
	)
	if parsed := net.ParseIP(host); parsed != nil {
		ip4 := parsed.To4()
		if ip4 == nil {
			return nil, errors.New("SOCKS4 does not support IPv6 addresses")
		}
		ip = ip4
	} else if d.Socks4A {
		// SOCKS4a: use invalid IP 0.0.0.x (x != 0) and append domain after userid.
		if host == "" {
			return nil, errors.New("SOCKS4a domain name is empty")
		}
		if len(host) > maxSocks4aDomainLen {
			return nil, errors.New("SOCKS4a domain name too long")
		}
		if err := validateSocks4CString(host, "domain name"); err != nil {
			return nil, err
		}
		ip = net.IPv4(0, 0, 0, 1).To4()
		domain = host
	} else {
		ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.New("no IPv4 address found for host")
		}
		ip = ips[0].To4()
		if ip == nil {
			return nil, errors.New("no IPv4 address found for host")
		}
	}

	// VN | CD | DSTPORT | DSTIP | USERID | NULL [| DOMAIN | NULL]
	b := make([]byte, 0, 9+len(d.UserID)+len(domain)+1)
	b = append(b, Version4, byte(d.cmd))
	b = append(b, byte(port>>8), byte(port))
	b = append(b, ip...)
	b = append(b, d.UserID...)
	b = append(b, 0)
	if domain != "" {
		b = append(b, domain...)
		b = append(b, 0)
	}
	if _, err := c.Write(b); err != nil {
		return nil, err
	}

	// Reply is always 8 bytes: VN | CD | DSTPORT | DSTIP
	var resp [8]byte
	if _, err := io.ReadFull(c, resp[:]); err != nil {
		return nil, err
	}
	// Spec says VN should be 0; some servers incorrectly echo 4.
	if resp[0] != 0 && resp[0] != Version4 {
		return nil, errors.New("unexpected protocol version " + strconv.Itoa(int(resp[0])))
	}
	if code := Reply(resp[1]); code != Status4Granted {
		return nil, errors.New(code.String())
	}

	a := &Addr{
		IP:   net.IPv4(resp[4], resp[5], resp[6], resp[7]),
		Port: int(resp[2])<<8 | int(resp[3]),
	}
	return a, nil
}

// validateSocks4CString rejects strings that would truncate a SOCKS4 C-string field.
func validateSocks4CString(s, field string) error {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return errors.New("invalid SOCKS4 " + field + ": contains NUL")
		}
	}
	return nil
}

// connect5 implements the SOCKS5 CONNECT handshake.
func (d *Dialer) connect5(ctx context.Context, c net.Conn, host string, port int) (net.Addr, error) {
	b := make([]byte, 0, 6+len(host)) // the size here is just an estimate
	b = append(b, Version5)
	if len(d.AuthMethods) == 0 || d.Authenticate == nil {
		b = append(b, 1, byte(AuthMethodNotRequired))
	} else {
		ams := d.AuthMethods
		if len(ams) > 255 {
			return nil, errors.New("too many authentication methods")
		}
		b = append(b, byte(len(ams)))
		for _, am := range ams {
			b = append(b, byte(am))
		}
	}
	if _, err := c.Write(b); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(c, b[:2]); err != nil {
		return nil, err
	}
	if b[0] != Version5 {
		return nil, errors.New("unexpected protocol version " + strconv.Itoa(int(b[0])))
	}
	am := AuthMethod(b[1])
	if am == AuthMethodNoAcceptableMethods {
		return nil, errors.New("no acceptable authentication methods")
	}
	if d.Authenticate != nil {
		if err := d.Authenticate(ctx, c, am); err != nil {
			return nil, err
		}
	}

	b = b[:0]
	b = append(b, Version5, byte(d.cmd), 0)
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			b = append(b, AddrTypeIPv4)
			b = append(b, ip4...)
		} else if ip6 := ip.To16(); ip6 != nil {
			b = append(b, AddrTypeIPv6)
			b = append(b, ip6...)
		} else {
			return nil, errors.New("unknown address type")
		}
	} else {
		if len(host) > 255 {
			return nil, errors.New("FQDN too long")
		}
		b = append(b, AddrTypeFQDN)
		b = append(b, byte(len(host)))
		b = append(b, host...)
	}
	b = append(b, byte(port>>8), byte(port))
	if _, err := c.Write(b); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(c, b[:4]); err != nil {
		return nil, err
	}
	if b[0] != Version5 {
		return nil, errors.New("unexpected protocol version " + strconv.Itoa(int(b[0])))
	}
	if cmdErr := Reply(b[1]); cmdErr != StatusSucceeded {
		return nil, errors.New("unknown error " + cmdErr.String())
	}
	if b[2] != 0 {
		return nil, errors.New("non-zero reserved field")
	}
	l := 2
	var a Addr
	switch b[3] {
	case AddrTypeIPv4:
		l += net.IPv4len
		a.IP = make(net.IP, net.IPv4len)
	case AddrTypeIPv6:
		l += net.IPv6len
		a.IP = make(net.IP, net.IPv6len)
	case AddrTypeFQDN:
		if _, err := io.ReadFull(c, b[:1]); err != nil {
			return nil, err
		}
		l += int(b[0])
	default:
		return nil, errors.New("unknown address type " + strconv.Itoa(int(b[3])))
	}
	if cap(b) < l {
		b = make([]byte, l)
	} else {
		b = b[:l]
	}
	if _, err := io.ReadFull(c, b); err != nil {
		return nil, err
	}
	if a.IP != nil {
		copy(a.IP, b)
	} else {
		a.Name = string(b[:len(b)-2])
	}
	a.Port = int(b[len(b)-2])<<8 | int(b[len(b)-1])
	return &a, nil
}

func splitHostPort(address string) (string, int, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}
	portnum, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}
	if 1 > portnum || portnum > 0xffff {
		return "", 0, errors.New("port number out of range " + port)
	}
	return host, portnum, nil
}
