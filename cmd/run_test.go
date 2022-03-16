package cmd

import (
	"os"
	"testing"

	ia "github.com/iter8-tools/iter8/action"
	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestRun(t *testing.T) {
	ia.SetupWithMock(t)

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)

	tests := []cmdTestCase{
		// run
		{
			name:   "run",
			cmd:    "run",
			golden: base.CompletePath("../testdata", "output/run.txt"),
		},
	}

	os.Chdir(base.CompletePath("../", "testdata"))
	runTestActionCmd(t, tests)
}
