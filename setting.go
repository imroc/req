package req

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

func getClient() *http.Client {
	if Client != nil {
		return Client
	}
	return defaultClient
}

func getTransport() *http.Transport {
	trans, _ := getClient().Transport.(*http.Transport)
	return trans
}

// EnableInsecureTLS
func EnableInsecureTLS(enable bool) {
	trans := getTransport()
	if trans == nil {
		return
	}
	if trans.TLSClientConfig == nil {
		trans.TLSClientConfig = &tls.Config{}
	}
	trans.TLSClientConfig.InsecureSkipVerify = enable
}

// EnableCookieenable or disable cookie manager
func EnableCookie(enable bool) {
	if enable {
		jar, _ := cookiejar.New(nil)
		getClient().Jar = jar
	} else {
		getClient().Jar = nil
	}
}

// SetTimeout sets the timeout for every request
func SetTimeout(d time.Duration) {
	getClient().Timeout = d
}

// SetProxyUrl set the simple proxy with fixed proxy url
func SetProxyUrl(rawurl string) error {
	trans := getTransport()
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

// SetProxy sets the proxy for every request
func SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	trans := getTransport()
	if trans == nil {
		return errors.New("req: no transport")
	}
	trans.Proxy = proxy
	return nil
}
