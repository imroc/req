package req

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// dumpConn is a net.Conn which writes to Writer and reads from Reader
type dumpConn struct {
	io.Writer
	io.Reader
}

func (c *dumpConn) Close() error                       { return nil }
func (c *dumpConn) LocalAddr() net.Addr                { return nil }
func (c *dumpConn) RemoteAddr() net.Addr               { return nil }
func (c *dumpConn) SetDeadline(t time.Time) error      { return nil }
func (c *dumpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *dumpConn) SetWriteDeadline(t time.Time) error { return nil }

// delegateReader is a reader that delegates to another reader,
// once it arrives on a channel.
type delegateReader struct {
	c chan io.Reader
	r io.Reader // nil until received from c
}

func (r *delegateReader) Read(p []byte) (int, error) {
	if r.r == nil {
		r.r = <-r.c
	}
	return r.r.Read(p)
}

func (r *Req) dumpRequest() ([]byte, error) {
	reqSend := new(http.Request)
	*reqSend = *r.req
	if reqSend.URL.Scheme == "https" {
		reqSend.URL = new(url.URL)
		*reqSend.URL = *r.req.URL
		reqSend.URL.Scheme = "http"
	}

	if reqSend.ContentLength > 0 && reqSend.ContentLength == int64(len(r.reqBody)) {
		reqSend.Body = ioutil.NopCloser(bytes.NewReader(r.reqBody))
	} else {
		reqSend.Body = nil
	}

	// Use the actual Transport code to record what we would send
	// on the wire, but not using TCP.  Use a Transport with a
	// custom dialer that returns a fake net.Conn that waits
	// for the full input (and recording it), and then responds
	// with a dummy response.
	var buf bytes.Buffer // records the output
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()
	dr := &delegateReader{c: make(chan io.Reader)}

	t := &http.Transport{
		Dial: func(net, addr string) (net.Conn, error) {
			return &dumpConn{io.MultiWriter(&buf, pw), dr}, nil
		},
	}
	defer t.CloseIdleConnections()

	clientDo := new(http.Client)
	*clientDo = *r.client
	clientDo.Transport = t

	// Wait for the request before replying with a dummy response:
	go func() {
		req, err := http.ReadRequest(bufio.NewReader(pr))
		if err == nil {
			// Ensure all the body is read; otherwise
			// we'll get a partial dump.
			io.Copy(ioutil.Discard, req.Body)
			req.Body.Close()
		}
		dr.c <- strings.NewReader("HTTP/1.1 204 No Content\r\nConnection: close\r\n\r\n")
	}()

	_, err := clientDo.Do(reqSend)
	if err != nil {
		return nil, err
	}
	buf.ReadBytes('\n')

	var b bytes.Buffer
	b.WriteString(strings.Join([]string{reqSend.Method, reqSend.URL.String(), reqSend.Proto}, " "))
	if ShowCost {
		b.WriteString(" " + r.cost.String())
	}
	b.WriteString("\r\n")
	b.Write(buf.Bytes())

	return b.Bytes(), nil
}

func (r *Req) dump() string {
	var buf bytes.Buffer
	reqDump, err := r.dumpRequest()
	if err != nil {
		buf.WriteString(err.Error())
	}
	if len(reqDump) > 0 {
		buf.Write(reqDump)
		if len(r.reqBody) > 0 {
			buf.WriteString("\r\n\r\n")
		}
		buf.WriteString("=================================")
	}

	respDump, err := httputil.DumpResponse(r.resp, false)
	if err != nil {
		buf.WriteString(err.Error())
	}
	if len(respDump) > 0 {
		buf.WriteString("\r\n\r\n")
		buf.Write(respDump)
		buf.Write(r.respBody)
	}
	return buf.String()
}
