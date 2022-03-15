package cmd

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
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

	os.Chdir(base.CompletePath("../testdata", "assertinputs"))
	runTestActionCmd(t, tests)
}
