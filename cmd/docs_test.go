package cmd

import (
	"fmt"
	"testing"
)

func TestDocs(t *testing.T) {
	tests := []cmdTestCase{
		// assert, SLOs
		{
			name: "create docs",
			cmd:  fmt.Sprintf("docs --commandDocsDir %v", t.TempDir()),
		},
	}

	runTestActionCmd(t, tests)
}
