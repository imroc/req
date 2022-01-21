package req

import (
	"fmt"
	"io"
	"os"
)

type DumpOptions struct {
	Output       io.Writer
	RequestHead  bool
	RequestBody  bool
	ResponseHead bool
	ResponseBody bool
	Async        bool
}

func (do *DumpOptions) Set(opts ...DumpOption) {
	for _, opt := range opts {
		opt(do)
	}
}

type DumpOption func(*DumpOptions)

func DumpAsync() DumpOption {
	return func(o *DumpOptions) {
		o.Async = true
	}
}

func DumpHead() DumpOption {
	return func(o *DumpOptions) {
		o.RequestHead = true
		o.ResponseHead = true
	}
}

func DumpBody() DumpOption {
	return func(o *DumpOptions) {
		o.RequestBody = true
		o.ResponseBody = true
	}
}

func DumpRequest() DumpOption {
	return func(o *DumpOptions) {
		o.RequestHead = true
		o.RequestBody = true
	}
}

func DumpResponse() DumpOption {
	return func(o *DumpOptions) {
		o.ResponseHead = true
		o.ResponseBody = true
	}
}

func DumpAll() DumpOption {
	return func(o *DumpOptions) {
		o.RequestHead = true
		o.RequestBody = true
		o.ResponseHead = true
		o.ResponseBody = true
	}
}

func DumpTo(output io.Writer) DumpOption {
	return func(o *DumpOptions) {
		o.Output = output
	}
}

func DumpToFile(filename string) DumpOption {
	return func(o *DumpOptions) {
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		o.Output = file
	}
}

func (d *dumper) WrapReadCloser(rc io.ReadCloser) io.ReadCloser {
	return &dumpReadCloser{rc, d}
}

type dumpReadCloser struct {
	io.ReadCloser
	dump *dumper
}

func (r *dumpReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.dump.dump(p[:n])
	if err == io.EOF {
		r.dump.dump([]byte("\r\n"))
	}
	return
}

func (d *dumper) WrapWriteCloser(rc io.WriteCloser) io.WriteCloser {
	return &dumpWriteCloser{rc, d}
}

type dumpWriteCloser struct {
	io.WriteCloser
	dump *dumper
}

func (w *dumpWriteCloser) Write(p []byte) (n int, err error) {
	n, err = w.WriteCloser.Write(p)
	w.dump.dump(p[:n])
	return
}

func (d *dumper) WrapReader(r io.Reader) io.Reader {
	return &dumpReader{
		r:    r,
		dump: d,
	}
}

type dumpReader struct {
	r    io.Reader
	dump *dumper
}

func (r *dumpReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.dump.dump(p[:n])
	return
}

type dumpWriter struct {
	w    io.Writer
	dump *dumper
}

func (w *dumpWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.dump.dump(p[:n])
	return
}

func (d *dumper) WrapWriter(w io.Writer) io.Writer {
	return &dumpWriter{
		w:    w,
		dump: d,
	}
}

type dumper struct {
	*DumpOptions
	ch chan []byte
}

func DefaultDumpOptions() *DumpOptions {
	return defaultDumpOptions
}

var defaultDumpOptions = &DumpOptions{
	Output:       os.Stdout,
	RequestBody:  true,
	ResponseBody: true,
	ResponseHead: true,
	RequestHead:  true,
}

func newDumper(opt *DumpOptions) *dumper {
	if opt == nil {
		opt = defaultDumpOptions
	}
	if opt.Output == nil {
		opt.Output = os.Stdout
	}
	d := &dumper{
		DumpOptions: opt,
		ch:          make(chan []byte, 20),
	}
	return d
}

func (d *dumper) dump(p []byte) {
	if len(p) == 0 {
		return
	}
	if d.Async {
		b := make([]byte, len(p))
		copy(b, p)
		d.ch <- b
		return
	}
	d.Output.Write(p)
}

func (d *dumper) Stop() {
	d.ch <- nil
}

func (d *dumper) Start() {
	for b := range d.ch {
		if b == nil {
			fmt.Println("stop dump")
			return
		}
		d.Output.Write(b)
	}
}

func (t *Transport) EnableDump(opt *DumpOptions) {
	dump := newDumper(opt)
	t.dump = dump
	go dump.Start()
}

func (t *Transport) DisableDump() {
	if t.dump != nil {
		t.dump.Stop()
		t.dump = nil
	}
}
