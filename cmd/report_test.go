package cmd

import (
	"os"
	"testing"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestReport(t *testing.T) {

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	tests := []cmdTestCase{
		// report text
		{
			name:   "report text",
			cmd:    "report",
			golden: base.CompletePath("../testdata", "output/report.txt"),
		},
	}

	os.Chdir(base.CompletePath("../testdata", "assertinputs"))
	runTestActionCmd(t, tests)
}
