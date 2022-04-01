package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestHub(t *testing.T) {
	tests := []cmdTestCase{
		// basic hub
		{
			name:   "basic hub",
			cmd:    "hub --gitFolder github.com/iter8-tools/iter8.git//chart",
			golden: base.CompletePath("../testdata", "output/hub.txt"),
		},
		// hub, chartsParentDir
		{
			name:   "hub with chartsParentDir",
			cmd:    fmt.Sprintf("hub --gitFolder github.com/iter8-tools/iter8.git//chart --chartsParentDir %v", t.TempDir()),
			golden: base.CompletePath("../testdata", "output/hub-with-destdir.txt"),
		},
	}

	os.Chdir(t.TempDir())
	runTestActionCmd(t, tests)
}
