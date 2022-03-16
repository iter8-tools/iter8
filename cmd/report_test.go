package cmd

import (
	"os"
	"testing"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

// Credit: this test structure is inspired by
// https://github.com/helm/helm/blob/main/cmd/helm/install_test.go
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
		// report HTML
		{
			name:   "report HTML",
			cmd:    "report -o html",
			golden: base.CompletePath("../testdata", "output/report.html"),
		},
	}

	os.Chdir(base.CompletePath("../testdata", "assertinputs"))
	runTestActionCmd(t, tests)
}
