package cmd

import (
	"testing"
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
