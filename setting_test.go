package req

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newDefaultTestServer() *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestSetClient(t *testing.T) {

	ts := newDefaultTestServer()

	client := &http.Client{}
	SetClient(client)
	_, err := Get(ts.URL)
	if err != nil {
		t.Errorf("error after set client: %v", err)
	}

	SetClient(nil)
	_, err = Get(ts.URL)
	if err != nil {
		t.Errorf("error after set client to nil: %v", err)
	}

	client = Client()
	if trans, ok := client.Transport.(*http.Transport); ok {
		trans.MaxIdleConns = 1
		trans.DisableKeepAlives = true
		_, err = Get(ts.URL)
		if err != nil {
			t.Errorf("error after change client's transport: %v", err)
		}
	} else {
		t.Errorf("transport is not http.Transport: %+#v", client.Transport)
	}
}

func TestSetting(t *testing.T) {
	defer func() {
		if rc := recover(); rc != nil {
			t.Errorf("panic happened while change setting: %v", rc)
		}
	}()
	SetTimeout(2 * time.Second)
	EnableCookie(false)
	EnableCookie(true)
	EnableInsecureTLS(true)
	SetJSONIndent("", "    ")
	SetJSONEscapeHTML(false)
	SetXMLIndent("", "\t")
	SetProxyUrl("http://localhost:8080")
	SetProxy(nil)
}
