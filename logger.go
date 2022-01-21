package req

import (
	"fmt"
	"io"
	"os"
)

type Logger interface {
	Println(v ...interface{})
}

type logger struct {
	w io.Writer
}

func (l *logger) Println(v ...interface{}) {
	fmt.Fprintln(l.w, v...)
}

func NewLogger(output io.Writer) Logger {
	if output == nil {
		output = os.Stdout
	}
	return &logger{output}
}

type emptyLogger struct{}

func (l *emptyLogger) Println(v ...interface{}) {}

func logp(logger Logger, s string) {
	logger.Println("[req]", s)
}

func logf(logger Logger, format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	logp(logger, s)
}
