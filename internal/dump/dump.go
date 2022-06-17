package dump

import (
	"context"
	"io"
	"net/http"
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

// GetResponseHeaderDumpers return Dumpers which need dump response header.
func GetResponseHeaderDumpers(ctx context.Context, dump *Dumper) Dumpers {
	dumpers := GetDumpers(ctx, dump)
	var ds []*Dumper
	for _, d := range dumpers {
		if d.ResponseHeader() {
			ds = append(ds, d)
		}
	}
	return Dumpers(ds)
}

// Dumpers is an array of Dumpper
type Dumpers []*Dumper

// ShouldDump is true if Dumper is not empty.
func (ds Dumpers) ShouldDump() bool {
	return len(ds) > 0
}

// Dump with all dumpers.
func (ds Dumpers) Dump(p []byte) {
	for _, d := range ds {
		d.Dump(p)
	}
}

// Dumper is the dump tool.
type Dumper struct {
	Options
	ch chan []byte
}

// NewDumper create a new Dumper.
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

func WrapResponseBodyIfNeeded(res *http.Response, req *http.Request, dump *Dumper) {
	dumps := GetDumpers(req.Context(), dump)
	for _, d := range dumps {
		if d.ResponseBody() {
			res.Body = d.WrapReadCloser(res.Body)
		}
	}
}
