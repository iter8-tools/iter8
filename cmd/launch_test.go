package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestLaunch(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// launch, chartsParentDir, noDownload
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c %v --localChart --set tasks={http} --set http.url=https://httpbin.org/get --set http.duration=2s", base.CompletePath("../charts", "iter8")),
			golden: base.CompletePath("../testdata", "output/launch.txt"),
		},
		// launch, values file
		{
			name:   "launch with values file",
			cmd:    fmt.Sprintf("launch -c %v --localChart --set tasks={http,assess} --set http.duration=2s -f %v", base.CompletePath("../charts", "iter8"), base.CompletePath("../testdata", "config.yaml")),
			golden: base.CompletePath("../testdata", "output/launch-with-slos.txt"),
		},
	}

	base.SetupWithMock(t)
	runTestActionCmd(t, tests)
}
