package main

import (
	"fmt"
	"io"
	"time"
	"github.com/imroc/req/v3"
)

type SlowReader struct {
	Size int
	n    int
}

func (r *SlowReader) Close() error {
	return nil
}

func (r *SlowReader) Read(p []byte) (int, error) {
	if r.n >= r.Size {
		return 0, io.EOF
	}
	time.Sleep(1 * time.Millisecond)
	n := len(p)
	if r.n+n >= r.Size {
		n = r.Size - r.n
	}
	for i := 0; i < n; i++ {
		p[i] = 'h'
	}
	r.n += n
	return n, nil
}

func main() {
	size := 10 * 1024 * 1024
	req.SetFileUpload(req.FileUpload{
		ParamName: "file",
		FileName:  "test.txt",
		GetFileContent: func() (io.ReadCloser, error) {
			return &SlowReader{Size: size}, nil
		},
		FileSize: int64(size),
	}).SetUploadCallbackWithInterval(func(info req.UploadInfo) {
		fmt.Printf("%s: %.2f%%\n", info.FileName, float64(info.UploadedSize)/float64(info.FileSize)*100.0)
	}, 30*time.Millisecond).Post("http://127.0.0.1:8888/upload")
}
