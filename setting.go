package req

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

var defaultCookieJar http.CookieJar

func init() {
	defaultCookieJar, _ = cookiejar.New(nil)
}

type setting struct {
	Timeout             time.Duration // total timeout
	DialTimeout         time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	TLSHandshakeTimeout time.Duration
	InsecureTLS         bool
	Jar                 http.CookieJar
	Proxy               func(*http.Request) (*url.URL, error)
	TlsClientConfig     *tls.Config
	Transport           *http.Transport
	Client              *http.Client
}

func (r *Request) prepareSetting() bool {
	if r == nil {
		return false
	}
	if r.setting == nil {
		r.setting = &setting{}
	}
	return true
}

// GetClient returns the *http.Client according to the setting.
func (r *Request) GetClient() *http.Client {
	if !r.prepareSetting() {
		return http.DefaultClient
	}
	s := r.setting
	if s.Client == nil {
		c := &http.Client{
			Transport: r.GetTransport(),
		}
		if s.Jar != nil {
			c.Jar = s.Jar
		}
		if s.Timeout > 0 {
			c.Timeout = s.Timeout
		}
		s.Client = c
	}
	return s.Client
}

func (r *Request) createTransport() *http.Transport {
	s := r.setting
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
	if s.TLSHandshakeTimeout > 0 {
		trans.TLSHandshakeTimeout = s.TLSHandshakeTimeout
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
func (r *Request) GetTransport() *http.Transport {
	if !r.prepareSetting() {
		trans, _ := http.DefaultTransport.(*http.Transport)
		return trans
	}
	s := r.setting
	if s.Transport == nil {
		s.Transport = r.createTransport()
	}
	return s.Transport
}

// Client set the http.Client for the request
func (r *Request) Client(client *http.Client) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.Client = client
	return r
}

// Transport set the http.Transport for request
func (r *Request) Transport(trans *http.Transport) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.Transport = trans
	r.setting.Client = nil
	return r
}

// Proxy set the proxy for request
func (r *Request) Proxy(proxy func(*http.Request) (*url.URL, error)) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.Proxy = proxy
	r.setting.Client = nil
	r.setting.Transport = nil
	return r
}

// Timeout set the total timeout for request,
// once timeout reached, request will be canceled.
func (r *Request) Timeout(d time.Duration) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.Timeout = d
	r.setting.Client = nil
	r.setting.Transport = nil
	return r
}

// TimeoutDial sets the timeout for dial connection.
func (r *Request) TimeoutDial(d time.Duration) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.DialTimeout = d
	r.setting.Client = nil
	r.setting.Transport = nil
	return r
}

// TimeoutRead sets the timeout for read operation.
func (r *Request) TimeoutRead(d time.Duration) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.ReadTimeout = d
	r.setting.Client = nil
	r.setting.Transport = nil
	return r
}

// TimeoutWrite sets the timeout for write operation.
func (r *Request) TimeoutWrite(d time.Duration) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.WriteTimeout = d
	r.setting.Client = nil
	r.setting.Transport = nil
	return r
}

// TimeoutTLSHandshake specifies the maximum amount of time waiting to
// wait for a TLS handshake. Zero means no timeout.
func (r *Request) TimeoutTLSHandshake(d time.Duration) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.TLSHandshakeTimeout = d
	r.setting.Client = nil
	r.setting.Transport = nil
	return r
}

// InsecureTLS allows to access the insecure https server.
func (r *Request) InsecureTLS(ins bool) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.InsecureTLS = ins
	r.setting.Client = nil
	r.setting.Transport = nil
	return r
}

// EnableCookie set the default CookieJar to the request if enable==true, otherwise set to nil.
func (r *Request) EnableCookie(enable bool) *Request {
	if !r.prepareSetting() {
		return nil
	}
	if enable {
		r.setting.Jar = defaultCookieJar
	} else {
		r.setting.Jar = nil
	}
	r.setting.Client = nil
	return r
}

// EnableCookieWithJar set the specified *http.CookieJar to the request.
func (r *Request) EnableCookieWithJar(jar http.CookieJar) *Request {
	if !r.prepareSetting() {
		return nil
	}
	r.setting.Jar = jar
	r.setting.Client = nil
	return r
}

// Merge clone some properties of another Request into this.
func (r *Request) Merge(rr *Request) *Request {
	if r == nil || rr == nil {
		return nil
	}
	if len(rr.params) > 0 { // merge params
		for name, value := range rr.params {
			if _, ok := r.params[name]; !ok {
				r.params[name] = value
			}
		}
	}
	if rr.req != nil { // merge internal http.Request
		if r.req == nil {
			r.req = basicRequest()
		}
		if r.req.Method == "" && rr.req.Method != "" {
			r.req.Method = rr.req.Method
		}
		if r.req.Host == "" && rr.req.Host != "" {
			r.req.Host = rr.req.Host
		}
		if r.req.Proto != rr.req.Proto {
			r.req.Proto = rr.req.Proto
			r.req.ProtoMajor = rr.req.ProtoMajor
			r.req.ProtoMinor = rr.req.ProtoMinor
		}
		for name, value := range rr.req.Header {
			if _, ok := r.req.Header[name]; !ok {
				r.req.Header[name] = value
			}
		}
	}
	if rr.setting != nil { // merge setting
		rr.GetClient() // ensure client has been created. prevent creating client every request.
		s := *rr.setting
		r.setting = &s
	}

	return r
}
