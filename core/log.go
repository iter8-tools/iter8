package core

import (
	"bufio"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

type StackTrace struct {
	Trace string
}

func (st StackTrace) String() string {
	out := "stack trace below ... \n"
	scanner := bufio.NewScanner(strings.NewReader(st.Trace))
	for scanner.Scan() {
		out += "::Trace:: " + scanner.Text() + "\n"
	}
	out = strings.TrimSuffix(out, "\n")
	return out
}

func init() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		DisableQuote:    true,
		DisableSorting:  true,
	})
}
