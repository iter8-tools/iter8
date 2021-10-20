package core

import (
	"bufio"
	"strings"

	"github.com/sirupsen/logrus"
)

// Iter8Logger inherits all methods from logrus logger
type Iter8Logger struct {
	*logrus.Logger
}

// StackTrace is the trace from external components like a shell scripts run by an Iter8 task
type StackTrace struct {
	Trace string
}

// Logger to be used in all of Iter8
var Logger *Iter8Logger

// Initialize logger
func init() {
	Logger = &Iter8Logger{logrus.New()}
	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		DisableQuote:    true,
		DisableSorting:  true,
	})
}

// WithStackTrace yields a log entry with a formatted stack trace field embedded in it
func (l *Iter8Logger) WithStackTrace(t string) *logrus.Entry {
	return l.WithField("stack-trace", &StackTrace{
		Trace: t,
	})
}

func (st *StackTrace) String() string {
	out := "stack trace below ... \n"
	scanner := bufio.NewScanner(strings.NewReader(st.Trace))
	for scanner.Scan() {
		out += "::Trace:: " + scanner.Text() + "\n"
	}
	out = strings.TrimSuffix(out, "\n")
	return out
}
