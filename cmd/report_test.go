package cmd

import (
	"os"
	"testing"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestReport(t *testing.T) {
	os.Chdir(t.TempDir())
	id.CopyFileToPwd(t, base.CompletePath("../testdata", "assertinputs/experiment.yaml"))
	tests := []cmdTestCase{
		// report text
		{
			name:   "report text",
			cmd:    "report",
			golden: base.CompletePath("../testdata", "output/report.txt"),
		},
	}

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	runTestActionCmd(t, tests)
}
