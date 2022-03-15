package cmd

import (
	"fmt"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"helm.sh/helm/v3/pkg/repo/repotest"
)

// Credit: this test structure is inspired by
// https://github.com/helm/helm/blob/main/cmd/helm/install_test.go
func TestLaunch(t *testing.T) {
	srv, err := repotest.NewTempServerWithCleanup(t, base.CompletePath("../", "testdata/charts/*.tgz*"))
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	if err := srv.LinkIndices(); err != nil {
		t.Fatal(err)
	}

	tests := []cmdTestCase{
		// Launch, base case, values from CLI
		{
			name:   "basic launch",
			cmd:    fmt.Sprintf("launch -c load-test-http --repoURL %v --set url=https://httpbin.org/get", srv.URL()),
			golden: base.CompletePath("../testdata", "output/launch.txt"),
		},
		// Launch, dest dir
		{
			name:   "install with destDir",
			cmd:    fmt.Sprintf("launch -c load-test-http --destDir %v --repoURL %v --set url=https://httpbin.org/get", t.TempDir(), srv.URL()),
			golden: base.CompletePath("../testdata", "output/launch-with-destdir.txt"),
		},
		// Launch, values file
		{
			name:   "launch with values file",
			cmd:    fmt.Sprintf("launch -c load-test-http --repoURL %v -f config.yaml", srv.URL()),
			golden: base.CompletePath("../testdata", "output/launch-with-values-file.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
