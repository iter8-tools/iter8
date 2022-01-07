// Package log provides primitives for logging.
package log

import (
	"bufio"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Iter8Logger inherits all methods from logrus logger.
type Iter8Logger struct {
	*logrus.Logger
}

// StackTrace is the trace from external components like a shell scripts run by an Iter8 task.
type StackTrace struct {
	// Trace is the raw trace
	Trace string
}

// Logger to be used in all of Iter8.
var Logger *Iter8Logger

// init initializes the logger.
func init() {
	Logger = &Iter8Logger{logrus.New()}
	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		DisableQuote:    true,
		DisableSorting:  true,
	})

	// initialize log level
	viper.BindEnv("LOG_LEVEL")
	viper.SetDefault("LOG_LEVEL", "info")
	ll, _ := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	Logger.Debug("LOG_LEVEL ", ll)
	SetLogLevel(ll)
}

// SetLogLevel to a given logrus log level.
func SetLogLevel(ll logrus.Level) {
	Logger.SetLevel(ll)
}

// WithStackTrace yields a log entry with a formatted stack trace field embedded in it.
func (l *Iter8Logger) WithStackTrace(t string) *logrus.Entry {
	return l.WithField("stack-trace", &StackTrace{
		Trace: t,
	})
}

// String processes the stack trace by prefixing each line of the trace with ::Trace::.
// This enables other tools like grep to easily filter out these traces if needed.
func (st *StackTrace) String() string {
	out := "below ... \n"
	scanner := bufio.NewScanner(strings.NewReader(st.Trace))
	for scanner.Scan() {
		out += "::Trace:: " + scanner.Text() + "\n"
	}
	out = strings.TrimSuffix(out, "\n")
	return out
}
