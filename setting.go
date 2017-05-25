package req

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Client return the default underlying http client
func (r *Req) Client() *http.Client {
	if r.client == nil {
		r.client = newClient()
	}
	return r.client
}

// Client return the default underlying http client
func Client() *http.Client {
	return std.Client()
}

// SetClient sets the underlying http.Client.
func (r *Req) SetClient(client *http.Client) {
	r.client = client // use default if client == nil
}

// SetClient sets the default http.Client for requests.
func SetClient(client *http.Client) {
	std.SetClient(client)
}

// SetFlags control display format of *Resp
func (r *Req) SetFlags(flags int) {
	r.flag = flags
}

// SetFlags control display format of *Resp
func SetFlags(flags int) {
	std.SetFlags(flags)
}

// Flags return output format for the *Resp
func (r *Req) Flags() int {
	return r.flag
}

// Flags return output format for the *Resp
func Flags() int {
	return std.Flags()
}

func (r *Req) getTransport() *http.Transport {
	trans, _ := r.client.Transport.(*http.Transport)
	return trans
}

// EnableInsecureTLS allows insecure https
func (r *Req) EnableInsecureTLS(enable bool) {
	trans := r.getTransport()
	if trans == nil {
		return
	}
	if trans.TLSClientConfig == nil {
		trans.TLSClientConfig = &tls.Config{}
	}
	trans.TLSClientConfig.InsecureSkipVerify = enable
}

func EnableInsecureTLS(enable bool) {
	std.EnableInsecureTLS(enable)
}

// EnableCookieenable or disable cookie manager
func (r *Req) EnableCookie(enable bool) {
	if enable {
		jar, _ := cookiejar.New(nil)
		r.client.Jar = jar
	} else {
		r.client.Jar = nil
	}
}

// EnableCookieenable or disable cookie manager
func EnableCookie(enable bool) {
	std.EnableCookie(enable)
}

// SetTimeout sets the timeout for every request
func (r *Req) SetTimeout(d time.Duration) {
	r.client.Timeout = d
}

// SetTimeout sets the timeout for every request
func SetTimeout(d time.Duration) {
	std.SetTimeout(d)
}

// SetProxyUrl set the simple proxy with fixed proxy url
func (r *Req) SetProxyUrl(rawurl string) error {
	trans := r.getTransport()
	if trans == nil {
		return errors.New("req: no transport")
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	trans.Proxy = http.ProxyURL(u)
	return nil
}

// SetProxyUrl set the simple proxy with fixed proxy url
func SetProxyUrl(rawurl string) error {
	return std.SetProxyUrl(rawurl)
}

// SetProxy sets the proxy for every request
func (r *Req) SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	trans := r.getTransport()
	if trans == nil {
		return errors.New("req: no transport")
	}
	trans.Proxy = proxy
	return nil
}

// SetProxy sets the proxy for every request
func SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	return std.SetProxy(proxy)
}
