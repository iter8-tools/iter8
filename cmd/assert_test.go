package cmd

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
)

func TestAssert(t *testing.T) {
	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)

	tests := []cmdTestCase{
		// assert, SLOs
		{
			name:   "assert SLOs",
			cmd:    "assert -c completed -c nofailure -c slos",
			golden: base.CompletePath("../testdata", "output/assert-slos.txt"),
		},
	}

	os.Chdir(base.CompletePath("../testdata", "assertinputs"))
	runTestActionCmd(t, tests)
}
