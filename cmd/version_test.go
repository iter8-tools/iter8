package cmd

import (
	"strings"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	tests := []cmdTestCase{
		// version
		{
			name: "version",
			cmd:  "version",
		},
		// version
		{
			name: "version short",
			cmd:  "version --short",
		},
	}

	runTestActionCmd(t, tests)

}

func TestVersionPrefix(t *testing.T) {
	assert.True(t, strings.HasPrefix(version, base.MajorMinor))
}
