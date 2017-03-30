package req

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

type Setting struct {
	Timeout         time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	InsecureTLS     bool
	Proxy           func(*http.Request) (*url.URL, error)
	TlsClientConfig *tls.Config
	Transport       http.RoundTripper
	Client          *http.Client
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

func (r *Request) SetTransport(trans http.RoundTripper) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.Transport = trans
	return r
}

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

func (r *Request) SetReadWriteTimeout(d time.Duration) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.ReadTimeout = d
	r.setting.WriteTimeout = d
	return r
}

// InsecureTLS insecure the https.
func (r *Request) SetInsecureTLS(ins bool) *Request {
	if r == nil {
		return nil
	}
	if r.setting == nil {
		r.setting = &Setting{}
	}
	r.setting.InsecureTLS = ins
	return r
}
