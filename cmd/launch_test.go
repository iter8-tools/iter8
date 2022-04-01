package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestLaunch(t *testing.T) {
	tests := []cmdTestCase{
		// launch, values from CLI
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c load-test-http --chartsParentDir %v --noDownload --set url=https://httpbin.org/get --set duration=2s", base.CompletePath("../", "")),
			golden: base.CompletePath("../testdata", "output/launch.txt"),
		},
		// launch, chartsParentDir
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c load-test-http --chartsParentDir %v --noDownload --runDir %v --set url=https://httpbin.org/get --set duration=2s", base.CompletePath("../", ""), t.TempDir()),
			golden: base.CompletePath("../testdata", "output/launch.txt"),
		},
		// launch, values file
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c load-test-http --chartsParentDir %v --noDownload --set duration=2s -f %v", base.CompletePath("../", ""), base.CompletePath("../testdata", "config.yaml")),
			golden: base.CompletePath("../testdata", "output/launch-with-slos.txt"),
		},
	}

	os.Chdir(t.TempDir())
	base.SetupWithMock(t)
	runTestActionCmd(t, tests)
}
