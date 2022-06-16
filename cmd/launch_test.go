package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
)

func TestLaunch(t *testing.T) {
	os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// launch, chartsParentDir
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c iter8 --chartsParentDir %v --noDownload --set tasks={http} --set http.url=https://httpbin.org/get --set http.duration=2s", base.CompletePath("../", "")),
			golden: base.CompletePath("../testdata", "output/launch.txt"),
		},
		// launch, values file
		{
			name:   "launch with values file",
			cmd:    fmt.Sprintf("launch -c iter8 --chartsParentDir %v --noDownload --set tasks={http,assess} --set http.duration=2s -f %v", base.CompletePath("../", ""), base.CompletePath("../testdata", "config.yaml")),
			golden: base.CompletePath("../testdata", "output/launch-with-slos.txt"),
		},
	}

	base.SetupWithMock(t)
	runTestActionCmd(t, tests)
}
