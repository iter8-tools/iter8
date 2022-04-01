package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestHub(t *testing.T) {
	os.Chdir(t.TempDir())

	tests := []cmdTestCase{
		// hub
		{
			name:   "basic hub",
			cmd:    fmt.Sprintf("hub -c load-test-http --repoURL %v"),
			golden: base.CompletePath("../testdata", "output/hub.txt"),
		},
		// hub, destDir
		{
			name:   "hub with destDir",
			cmd:    fmt.Sprintf("hub -c load-test-http --destDir %v --repoURL %v", t.TempDir()),
			golden: base.CompletePath("../testdata", "output/hub-with-destdir.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
