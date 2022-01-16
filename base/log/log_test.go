package log

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackTrace(t *testing.T) {
	st := StackTrace{
		Trace: fmt.Sprintln("a") + fmt.Sprintln("b"),
	}
	assert.Contains(t, st.String(), "::Trace:: a")
	assert.Contains(t, st.String(), "::Trace:: b")
}
