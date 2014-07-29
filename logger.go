package ninja

import (
	"github.com/bugsnag/bugsnag-go"
	"github.com/juju/loggo"
)

// Logger wrapper for the internal logger with some extra helpers
type Logger struct {
	loggo.Logger
}

// GetLogger builds a ninja logger with the given name
func GetLogger(name string) *Logger {
	l := loggo.GetLogger(name)
	l.SetLogLevel(loggo.INFO)
	return &Logger{l}
}

// HandleError This notifies bugsnag and logs the error.
func (l *Logger) HandleError(err error, msg string) {
	l.Errorf("%s : %v", msg, err)
	bugsnag.Notify(err)
}
