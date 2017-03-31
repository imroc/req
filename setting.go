package req

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Setting struct {
	Timeout             time.Duration // total timeout
	DialTimeout         time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	TLSHandshakeTimeout time.Duration
	InsecureTLS         bool
	Proxy               func(*http.Request) (*url.URL, error)
	TlsClientConfig     *tls.Config
	Transport           *http.Transport
	Client              *http.Client
}

// GetClient returns the http.Client according to the setting.
func (s *Setting) GetClient() *http.Client {
	if s == nil {
		return http.DefaultClient
	}
	if s.Client == nil {
		c := &http.Client{
			Transport: s.GetTransport(),
		}
		if s.Timeout > 0 {
			c.Timeout = s.Timeout
		}
		s.Client = c
	}
	return s.Client
}

func (s *Setting) createTransport() *http.Transport {
	trans := &http.Transport{}
	trans.Dial = func(network, address string) (conn net.Conn, err error) {
		if s.DialTimeout > 0 {
			conn, err = net.DialTimeout(network, address, s.DialTimeout)
			if err != nil {
				return
			}
		} else {
			conn, err = net.Dial(network, address)
			if err != nil {
				return
			}
		}
		if s.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
		}
		if s.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout))
		}
		return
	}
	if s.TlsClientConfig != nil {
		trans.TLSClientConfig = s.TlsClientConfig
	}
	if s.InsecureTLS {
		if trans.TLSClientConfig != nil {
			trans.TLSClientConfig.InsecureSkipVerify = true
		} else {
			trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}
	if s.Proxy != nil {
		trans.Proxy = s.Proxy
	} else {
		trans.Proxy = http.ProxyFromEnvironment
	}
	return trans
}

// GetTransport return the http.Transport according to the setting.
func (s *Setting) GetTransport() *http.Transport {
	if s == nil {
		trans, _ := http.DefaultTransport.(*http.Transport)
		return trans
	}
	if s.Transport == nil {
		s.Transport = s.createTransport()
	}
	return s.Transport
}

func (r *Request) SetClient(client *http.Client) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.Client = client
	return r
}

// SetTransport set the http.Transport for request
func (r *Request) SetTransport(trans *http.Transport) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.Transport = trans
	return r
}

// SetProxy set the proxy for request
func (r *Request) SetProxy(proxy func(*http.Request) (*url.URL, error)) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.Proxy = proxy
	return r
}

// SetTimeout set the total timeout for request,
// once timeout reached, request will be canceled.
func (r *Request) SetTimeout(d time.Duration) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.Timeout = d
	return r
}

// SetDialTimeout sets the timeout for dial connection.
func (r *Request) SetDialTimeout(d time.Duration) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.DialTimeout = d
	return r
}

// SetReadTimeout sets the timeout for read operation.
func (r *Request) SetReadTimeout(d time.Duration) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.ReadTimeout = d
	return r
}

// SetWriteTimeout sets the timeout for write operation.
func (r *Request) SetWriteTimeout(d time.Duration) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.WriteTimeout = d
	return r
}

// EnableInsecureTLS insecure the https.
func (r *Request) EnableInsecureTLS(ins bool) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.InsecureTLS = ins
	return r
}
