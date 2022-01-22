package req

import (
	"fmt"
	"io"
	"os"
)

// Logger is the logging interface that req used internal,
// you can set the Logger for client if you want to see req's
// internal logging information.
type Logger interface {
	Println(v ...interface{})
}

type logger struct {
	w io.Writer
}

func (l *logger) Println(v ...interface{}) {
	fmt.Fprintln(l.w, v...)
}

// NewLogger create a simple Logger.
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
