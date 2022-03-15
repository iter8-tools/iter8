package cmd

import (
	"testing"
)

// Credit: this test structure is inspired by
// https://github.com/helm/helm/blob/main/cmd/helm/install_test.go
func TestVersion(t *testing.T) {
	tests := []cmdTestCase{
		// version
		{
			name: "version",
			cmd:  "version",
		},
	}

	runTestActionCmd(t, tests)

}
