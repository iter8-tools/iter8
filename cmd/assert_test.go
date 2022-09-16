package cmd

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
)

func TestAssert(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	_ = id.CopyFileToPwd(t, base.CompletePath("../testdata", "assertinputs/experiment.yaml"))
	tests := []cmdTestCase{
		// assert, SLOs
		{
			name:   "assert SLOs",
			cmd:    "assert -c completed -c nofailure -c slos",
			golden: base.CompletePath("../testdata", "output/assert-slos.txt"),
		},
	}

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	runTestActionCmd(t, tests)
}
