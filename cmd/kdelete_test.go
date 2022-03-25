package cmd

import (
	"fmt"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
)

func TestKDelete(t *testing.T) {
	srv := id.SetupWithRepo(t)
	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)

	tests := []cmdTestCase{
		// Launch, base case, values from CLI
		{
			name: "basic k launch",
			cmd:  fmt.Sprintf("k launch -c load-test-http --repoURL %v --set url=https://httpbin.org/get --set duration=2s", srv.URL()),
		},
		// Launch again, values from CLI
		{
			name: "launch again",
			cmd:  fmt.Sprintf("k launch -c load-test-http --repoURL %v --set url=https://httpbin.org/get --set duration=2s", srv.URL()),
		},
		// Delete
		{
			name:   "delete",
			cmd:    "k delete",
			golden: base.CompletePath("../testdata", "output/kdelete.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
