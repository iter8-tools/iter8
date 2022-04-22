package cmd

import (
	"fmt"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
)

func TestKDelete(t *testing.T) {
	tests := []cmdTestCase{
		// Launch, base case, values from CLI
		{
			name:   "basic k launch",
			cmd:    fmt.Sprintf("k launch -c load-test-http --chartsParentDir %v --noDownload --set url=https://httpbin.org/get --set duration=2s", base.CompletePath("../", "")),
			golden: base.CompletePath("../testdata", "output/klaunch.txt"),
		},
		// Launch again, values from CLI
		{
			name: "launch again",
			cmd:  fmt.Sprintf("k launch -c load-test-http --chartsParentDir %v --noDownload --set url=https://httpbin.org/get --set duration=2s", base.CompletePath("../", "")),
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
