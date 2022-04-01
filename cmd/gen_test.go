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
			cmd:    fmt.Sprintf("gen --sourceDir %v --set url=https://httpbin.org", base.CompletePath("../testdata", "charts/load-test-http")),
			golden: base.CompletePath("../testdata", "output/gen-cli-values.txt"),
		},
		// gen, values file
		{
			name:   "gen with values file",
			cmd:    fmt.Sprintf("gen --sourceDir %v --set duration=2s -f %v", base.CompletePath("../testdata", "charts/load-test-http"), base.CompletePath("../testdata", "config.yaml")),
			golden: base.CompletePath("../testdata", "output/gen-values-file.txt"),
		},
	}

	os.Chdir(t.TempDir())
	runTestActionCmd(t, tests)
}
