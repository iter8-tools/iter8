package log

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackTrace(t *testing.T) {
	Logger.WithIndentedTrace("hello there")
	Logger.WithStackTrace("hello there")

	st := StackTrace{
		prefix: "::Trace:: ",
		Trace:  fmt.Sprintln("a") + fmt.Sprintln("b"),
	}
	assert.Contains(t, st.String(), "::Trace:: a")
	assert.Contains(t, st.String(), "::Trace:: b")

}
