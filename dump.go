package req

import (
	"context"
	"io"
	"os"
)

// DumpOptions controls the dump behavior.
type DumpOptions struct {
	Output         io.Writer
	RequestHeader  bool
	RequestBody    bool
	ResponseHeader bool
	ResponseBody   bool
	Async          bool
}

func (do *DumpOptions) Clone() *DumpOptions {
	if do == nil {
		return nil
	}
	d := *do
	return &d
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

func newDefaultDumpOptions() *DumpOptions {
	return &DumpOptions{
		Output:         os.Stdout,
		RequestBody:    true,
		ResponseBody:   true,
		ResponseHeader: true,
		RequestHeader:  true,
	}
}

func newDumper(opt *DumpOptions) *dumper {
	if opt == nil {
		opt = newDefaultDumpOptions()
	}
	if opt.Output == nil {
		opt.Output = os.Stderr
	}
	d := &dumper{
		DumpOptions: opt,
		ch:          make(chan []byte, 20),
	}
	return d
}

func (d *dumper) Clone() *dumper {
	if d == nil {
		return nil
	}
	return &dumper{
		DumpOptions: d.DumpOptions.Clone(),
		ch:          make(chan []byte, 20),
	}
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
			return
		}
		d.Output.Write(b)
	}
}

func getDumperOverride(dump *dumper, ctx context.Context) *dumper {
	if d, ok := ctx.Value("dumper").(*dumper); ok {
		return d
	}
	return dump
}
