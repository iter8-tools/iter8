package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestDocs(t *testing.T) {
	os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// assert, SLOs
		{
			name: "create docs",
			cmd:  fmt.Sprintf("docs --commandDocsDir %v", t.TempDir()),
		},
	}

	runTestActionCmd(t, tests)
}
