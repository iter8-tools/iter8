package cmd

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestHub(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// basic hub
		{
			name:   "basic hub",
			cmd:    "hub --remoteFolderURL github.com/iter8-tools/iter8.git//charts",
			golden: base.CompletePath("../testdata", "output/hub.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
