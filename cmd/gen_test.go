package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestGen(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// gen, with CLI values
		{
			name:   "gen with CLI values",
			cmd:    fmt.Sprintf("gen -c %v --set tasks={http} --set http.url=https://httpbin.org", base.CompletePath("../charts", "iter8")),
			golden: base.CompletePath("../testdata", "output/gen-cli-values.txt"),
		},
		// gen, values file
		{
			name:   "gen with values file",
			cmd:    fmt.Sprintf("gen -c %v --set tasks={http,assess} --set http.duration=2s -f %v", base.CompletePath("../charts", "iter8"), base.CompletePath("../testdata", "config.yaml")),
			golden: base.CompletePath("../testdata", "output/gen-values-file.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
