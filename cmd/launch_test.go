package cmd

import (
	"fmt"
	"os"
	"testing"

	ia "github.com/iter8-tools/iter8/action"

	"github.com/iter8-tools/iter8/base"
)

func TestLaunch(t *testing.T) {
	srv := ia.SetupWithRepo(t)

	tests := []cmdTestCase{
		// launch, values from CLI
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c load-test-http --repoURL %v --set url=https://httpbin.org/get --set duration=2s", srv.URL()),
			golden: base.CompletePath("../testdata", "output/launch.txt"),
		},
		// launch, destDir
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c load-test-http --destDir %v --repoURL %v --set url=https://httpbin.org/get --set duration=2s", t.TempDir(), srv.URL()),
			golden: base.CompletePath("../testdata", "output/launch-with-destdir.txt"),
		},
		// launch, values file
		{
			name:   "launch with values file",
			cmd:    fmt.Sprintf("launch -c load-test-http --repoURL %v --set duration=2s -f %v", srv.URL(), base.CompletePath("../testdata", "config.yaml")),
			golden: base.CompletePath("../testdata", "output/launch-with-values-file.txt"),
		},
	}

	os.Chdir(t.TempDir())
	base.SetupWithMock(t)
	runTestActionCmd(t, tests)

}
