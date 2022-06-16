package cmd

import (
	"os"
	"testing"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestRun(t *testing.T) {
	os.Chdir(t.TempDir())
	id.CopyFileToPwd(t, base.CompletePath("../", "testdata/experiment.yaml"))
	base.SetupWithMock(t)

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

	runTestActionCmd(t, tests)
}
