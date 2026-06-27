package dump

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

// testOptions implements Options for testing
type testOptions struct {
	output            io.Writer
	requestHeader     bool
	requestBody       bool
	responseHeader    bool
	responseBody      bool
	async             bool
	requestHeaderOut  io.Writer
	requestBodyOut    io.Writer
	responseHeaderOut io.Writer
	responseBodyOut   io.Writer
}

func (o *testOptions) Output() io.Writer             { return o.output }
func (o *testOptions) RequestHeaderOutput() io.Writer { return o.requestHeaderOut }
func (o *testOptions) RequestBodyOutput() io.Writer   { return o.requestBodyOut }
func (o *testOptions) ResponseHeaderOutput() io.Writer { return o.responseHeaderOut }
func (o *testOptions) ResponseBodyOutput() io.Writer  { return o.responseBodyOut }
func (o *testOptions) RequestHeader() bool            { return o.requestHeader }
func (o *testOptions) RequestBody() bool              { return o.requestBody }
func (o *testOptions) ResponseHeader() bool           { return o.responseHeader }
func (o *testOptions) ResponseBody() bool             { return o.responseBody }
func (o *testOptions) Async() bool                    { return o.async }
func (o *testOptions) Clone() Options {
	return &testOptions{
		output:            o.output,
		requestHeader:     o.requestHeader,
		requestBody:       o.requestBody,
		responseHeader:    o.responseHeader,
		responseBody:      o.responseBody,
		async:             o.async,
		requestHeaderOut:  o.requestHeaderOut,
		requestBodyOut:    o.requestBodyOut,
		responseHeaderOut: o.responseHeaderOut,
		responseBodyOut:   o.responseBodyOut,
	}
}

func TestDumperSync(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:         &buf,
		requestHeader:  true,
		requestHeaderOut: &buf,
	}
	d := NewDumper(opt)
	d.DumpRequestHeader([]byte("GET / HTTP/1.1\r\n"))
	if buf.String() != "GET / HTTP/1.1\r\n" {
		t.Fatalf("expected 'GET / HTTP/1.1\\r\\n', got %q", buf.String())
	}
}

func TestDumperAsync(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:            &buf,
		requestHeader:     true,
		async:             true,
		requestHeaderOut:  &buf,
	}
	d := NewDumper(opt)
	go d.Start()
	defer d.Stop()

	d.DumpRequestHeader([]byte("GET / HTTP/1.1\r\n"))
	// Wait for async processing
	for i := 0; i < 100 && buf.Len() == 0; i++ {
		d.ch <- nil // trigger check
		time.Sleep(time.Millisecond)
	}
	// Give it a moment to process
	time.Sleep(50 * time.Millisecond)
	if buf.String() != "GET / HTTP/1.1\r\n" {
		t.Fatalf("expected 'GET / HTTP/1.1\\r\\n', got %q", buf.String())
	}
}

func TestDumperClone(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:           &buf,
		requestHeader:    true,
		requestHeaderOut: &buf,
	}
	d := NewDumper(opt)
	cloned := d.Clone()
	if cloned == nil {
		t.Fatal("clone returned nil")
	}
	if !cloned.RequestHeader() {
		t.Fatal("clone should have RequestHeader=true")
	}
}

func TestDumperCloneNil(t *testing.T) {
	var d *Dumper
	if d.Clone() != nil {
		t.Fatal("nil clone should return nil")
	}
}

func TestDumperEmptyData(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:         &buf,
		requestHeaderOut: &buf,
	}
	d := NewDumper(opt)
	d.DumpRequestHeader([]byte{})
	if buf.Len() != 0 {
		t.Fatal("expected no output for empty data")
	}
}

func TestDumpResponseBodyReadCloser(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:          &buf,
		responseBody:    true,
		responseBodyOut: &buf,
	}
	d := NewDumper(opt)

	body := io.NopCloser(bytes.NewReader([]byte("hello world")))
	wrapped := d.WrapResponseBodyReadCloser(body)

	buf2 := make([]byte, 11)
	n, err := wrapped.Read(buf2)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 11 {
		t.Fatalf("expected 11 bytes, got %d", n)
	}
	if buf.String() != "hello world" {
		t.Fatalf("expected 'hello world' in dump, got %q", buf.String())
	}
}

func TestDumpRequestBodyWriteCloser(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:         &buf,
		requestBody:    true,
		requestBodyOut: &buf,
	}
	d := NewDumper(opt)

	var inner bytes.Buffer
	wc := &nopWriteCloser{&inner}
	wrapped := d.WrapRequestBodyWriteCloser(wc)

	_, err := wrapped.Write([]byte("request body"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "request body" {
		t.Fatalf("expected 'request body' in dump, got %q", buf.String())
	}
	if inner.String() != "request body" {
		t.Fatalf("expected 'request body' in inner writer, got %q", inner.String())
	}
}

func TestGetDumpersWithContext(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{output: &buf}
	d := NewDumper(opt)

	// Without context dumper
	dumps := GetDumpers(context.Background(), d)
	if len(dumps) != 1 {
		t.Fatalf("expected 1 dumper, got %d", len(dumps))
	}

	// With context dumper
	ctx := context.WithValue(context.Background(), DumperKey, NewDumper(opt))
	dumps = GetDumpers(ctx, d)
	if len(dumps) != 2 {
		t.Fatalf("expected 2 dumpers, got %d", len(dumps))
	}

	// Nil dump arg
	dumps = GetDumpers(ctx, nil)
	if len(dumps) != 1 {
		t.Fatalf("expected 1 dumper from context, got %d", len(dumps))
	}
}

func TestGetResponseHeaderDumpers(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:           &buf,
		responseHeader:   true,
		responseHeaderOut: &buf,
	}
	d := NewDumper(opt)

	dumps := GetResponseHeaderDumpers(context.Background(), d)
	if !dumps.ShouldDump() {
		t.Fatal("expected ShouldDump to be true")
	}
	if len(dumps) != 1 {
		t.Fatalf("expected 1 dumper, got %d", len(dumps))
	}

	// Test with non-response-header dumper
	opt2 := &testOptions{output: &buf, responseHeader: false}
	d2 := NewDumper(opt2)
	dumps2 := GetResponseHeaderDumpers(context.Background(), d2)
	if dumps2.ShouldDump() {
		t.Fatal("expected ShouldDump to be false")
	}
}

func TestDumpersShouldDump(t *testing.T) {
	var ds Dumpers
	if ds.ShouldDump() {
		t.Fatal("empty Dumpers should not ShouldDump")
	}

	ds = Dumpers{NewDumper(&testOptions{output: io.Discard})}
	if !ds.ShouldDump() {
		t.Fatal("non-empty Dumpers should ShouldDump")
	}
}

func TestWrapResponseBodyIfNeeded(t *testing.T) {
	var buf bytes.Buffer
	opt := &testOptions{
		output:          &buf,
		responseBody:    true,
		responseBodyOut: &buf,
	}
	d := NewDumper(opt)

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	res := &http.Response{
		Body: io.NopCloser(bytes.NewReader([]byte("test"))),
	}
	WrapResponseBodyIfNeeded(res, req, d)

	// Read from the wrapped body
	buf2 := make([]byte, 4)
	n, _ := res.Body.Read(buf2)
	if n != 4 {
		t.Fatalf("expected 4 bytes, got %d", n)
	}
	if buf.String() != "test" {
		t.Fatalf("expected 'test' in dump, got %q", buf.String())
	}
}

// nopWriteCloser wraps a writer to implement io.WriteCloser
type nopWriteCloser struct {
	io.Writer
}

func (n *nopWriteCloser) Close() error { return nil }
