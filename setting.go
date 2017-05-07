package req

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
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
