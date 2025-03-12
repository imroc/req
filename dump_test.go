package req

import (
	"bytes"
	"fmt"
	"github.com/imroc/req/v3/internal/tests"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hjson/hjson-go/v4"
)

type hjsonFormatter struct {
	buf bytes.Buffer
}

func (d *hjsonFormatter) BodyFormat(p []byte, header http.Header) (formatted []byte, dumpImmediately bool) {
	if !strings.HasPrefix(header.Get("Content-Type"), "application/json") {
		return p, true
	}

	d.buf.Write(p)

	buf := d.buf.Bytes()
	var data interface{}
	if err := hjson.Unmarshal(buf, &data); err != nil {
		return nil, false
	}

	d.buf.Reset()

	hb, err := hjson.Marshal(data)
	if err != nil {
		return []byte(fmt.Sprintf("hjson error: %v", err)), true
	}

	return hb, true
}

func TestBodyFormatter(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Date", "Wed, 12 Mar 2025 05:39:45 GMT")
		w.Write([]byte(`{"name": "bingoohuang", "age": 134}`))
	}))
	server.Start()
	defer server.Close()

	c := C()
	opt := NewDefaultDumpOptions()
	var buf bytes.Buffer
	opt.Output = &buf
	opt.RequestBodyFormat = &hjsonFormatter{}
	opt.ResponseBodyFormat = &hjsonFormatter{}
	c.EnableDump(opt)

	c.R().
		SetHeader("Host", "localhost").
		SetBody(map[string]any{"highhandedly": "gastroduodenoscopy", "epipodite": 13.4}).
		Post(server.URL)

	expected := `POST / HTTP/1.1
Host: localhost
User-Agent: req/v3 (https://github.com/imroc/req)
Content-Length: 54
Content-Type: application/json; charset=utf-8
Accept-Encoding: gzip

{
  epipodite: 13.4
  highhandedly: gastroduodenoscopy
}
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 12 Mar 2025 05:39:45 GMT
Content-Length: 35

{
  age: 134
  name: bingoohuang
}
`

	tests.AssertEqual(t, expected, strings.ReplaceAll(buf.String(), "\r", ""))
}
