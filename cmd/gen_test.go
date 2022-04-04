package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestGen(t *testing.T) {
	tests := []cmdTestCase{
		// gen, with CLI values
		{
			name:   "gen with CLI values",
			cmd:    fmt.Sprintf("gen -c load-test-http --chartsParentDir %v --set url=https://httpbin.org", base.CompletePath("../", "")),
			golden: base.CompletePath("../testdata", "output/gen-cli-values.txt"),
		},
		// gen, values file
		{
			name:   "gen with values file",
			cmd:    fmt.Sprintf("gen -c load-test-http --chartsParentDir %v --set duration=2s -f %v", base.CompletePath("../", ""), base.CompletePath("../testdata", "config.yaml")),
			golden: base.CompletePath("../testdata", "output/gen-values-file.txt"),
		},
	}

	os.Chdir(t.TempDir())
	runTestActionCmd(t, tests)
}
