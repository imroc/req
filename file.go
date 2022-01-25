package req

import "io"

type uploadFile struct {
	ParamName string
	FilePath  string
	io.Reader
}
