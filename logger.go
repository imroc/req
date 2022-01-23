package req

import (
	"log"
	"os"
)

// Logger interface is to abstract the logging from Resty. Gives control to
// the Resty users, choice of the logger.
type Logger interface {
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

func createLogger() *logger {
	l := &logger{l: log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)}
	return l
}

var _ Logger = (*logger)(nil)

type disableLogger struct{}

func (l *disableLogger) Errorf(format string, v ...interface{}) {}
func (l *disableLogger) Warnf(format string, v ...interface{})  {}
func (l *disableLogger) Debugf(format string, v ...interface{}) {}

type logger struct {
	l *log.Logger
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.output("ERROR", format, v...)
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.output("WARN", format, v...)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.output("DEBUG", format, v...)
}

func (l *logger) output(level, format string, v ...interface{}) {
	format = level + " [req] " + format
	if len(v) == 0 {
		l.l.Print(format)
		return
	}
	l.l.Printf(format, v...)
}
