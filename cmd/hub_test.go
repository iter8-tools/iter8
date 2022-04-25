package cmd

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestHub(t *testing.T) {
	tests := []cmdTestCase{
		// basic hub
		{
			name:   "basic hub",
			cmd:    "hub --folder github.com/iter8-tools/iter8.git//charts",
			golden: base.CompletePath("../testdata", "output/hub.txt"),
		},
	}

	os.Chdir(t.TempDir())
	runTestActionCmd(t, tests)
}
