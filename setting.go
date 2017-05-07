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
func Client() *http.Client {
	return defaultClient
}

// SetClient sets the default http.Client for requests.
func SetClient(client *http.Client) {
	if client != nil {
		defaultClient = client
	}
}

func getTransport() *http.Transport {
	trans, _ := defaultClient.Transport.(*http.Transport)
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
		defaultClient.Jar = jar
	} else {
		defaultClient.Jar = nil
	}
}

// SetTimeout sets the timeout for every request
func SetTimeout(d time.Duration) {
	defaultClient.Timeout = d
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
