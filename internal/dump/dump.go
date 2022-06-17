package dump

import (
	"context"
	"io"
)

// Options controls the dump behavior.
type Options interface {
	Output() io.Writer
	RequestHeader() bool
	RequestBody() bool
	ResponseHeader() bool
	ResponseBody() bool
	Async() bool
	Clone() Options
}

func (d *Dumper) WrapReadCloser(rc io.ReadCloser) io.ReadCloser {
	return &dumpReadCloser{rc, d}
}

type dumpReadCloser struct {
	io.ReadCloser
	dump *Dumper
}

func (r *dumpReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.dump.Dump(p[:n])
	if err == io.EOF {
		r.dump.Dump([]byte("\r\n"))
	}
	return
}

func (d *Dumper) WrapWriteCloser(rc io.WriteCloser) io.WriteCloser {
	return &dumpWriteCloser{rc, d}
}

type dumpWriteCloser struct {
	io.WriteCloser
	dump *Dumper
}

func (w *dumpWriteCloser) Write(p []byte) (n int, err error) {
	n, err = w.WriteCloser.Write(p)
	w.dump.Dump(p[:n])
	return
}

type dumpWriter struct {
	w    io.Writer
	dump *Dumper
}

func (w *dumpWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.dump.Dump(p[:n])
	return
}

func (d *Dumper) WrapWriter(w io.Writer) io.Writer {
	return &dumpWriter{
		w:    w,
		dump: d,
	}
}

type Dumper struct {
	Options
	ch chan []byte
}

func NewDumper(opt Options) *Dumper {
	d := &Dumper{
		Options: opt,
		ch:      make(chan []byte, 20),
	}
	return d
}

func (d *Dumper) SetOptions(opt Options) {
	d.Options = opt
	return
}

func (d *Dumper) Clone() *Dumper {
	if d == nil {
		return nil
	}
	return &Dumper{
		Options: d.Options.Clone(),
		ch:      make(chan []byte, 20),
	}
}

func (d *Dumper) Dump(p []byte) {
	if len(p) == 0 {
		return
	}
	if d.Async() {
		b := make([]byte, len(p))
		copy(b, p)
		d.ch <- b
		return
	}
	d.Output().Write(p)
}

func (d *Dumper) Stop() {
	d.ch <- nil
}

func (d *Dumper) Start() {
	for b := range d.ch {
		if b == nil {
			return
		}
		d.Output().Write(b)
	}
}

type dumperKeyType int

const DumperKey dumperKeyType = iota

func GetDumpers(ctx context.Context, dump *Dumper) []*Dumper {
	dumps := []*Dumper{}
	if dump != nil {
		dumps = append(dumps, dump)
	}
	if d, ok := ctx.Value(DumperKey).(*Dumper); ok {
		dumps = append(dumps, d)
	}
	return dumps
}
