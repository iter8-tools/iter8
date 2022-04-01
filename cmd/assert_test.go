package cmd

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
)

func TestAssert(t *testing.T) {
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
	os.Chdir(base.CompletePath("../testdata", "assertinputs"))
	runTestActionCmd(t, tests)
}
