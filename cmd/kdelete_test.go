package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
)

func TestKDelete(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// Launch, base case, values from CLI
		{
			name:   "basic k launch",
			cmd:    fmt.Sprintf("k launch -c iter8 --chartsParentDir %v --noDownload --set tasks={http} --set http.url=https://httpbin.org/get --set http.duration=2s", base.CompletePath("../", "")),
			golden: base.CompletePath("../testdata", "output/klaunch.txt"),
		},
		// Launch again, values from CLI
		{
			name: "launch again",
			cmd:  fmt.Sprintf("k launch -c iter8 --chartsParentDir %v --noDownload --set tasks={http} --set http.url=https://httpbin.org/get --set http.duration=2s", base.CompletePath("../", "")),
		},
		// Delete
		{
			name:   "delete",
			cmd:    "k delete",
			golden: base.CompletePath("../testdata", "output/kdelete.txt"),
		},
	}

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	runTestActionCmd(t, tests)
}
