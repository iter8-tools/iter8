package cmd

import (
	"os"
	"testing"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestReport(t *testing.T) {
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
	os.Chdir(base.CompletePath("../testdata", "assertinputs"))
	runTestActionCmd(t, tests)
}
