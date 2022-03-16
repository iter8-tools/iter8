package cmd

import (
	"fmt"
	"testing"

	ia "github.com/iter8-tools/iter8/action"

	"github.com/iter8-tools/iter8/base"
)

func TestKLaunch(t *testing.T) {
	srv := ia.SetupWithRepo(t)

	tests := []cmdTestCase{
		// Launch, base case, values from CLI
		{
			name:   "basic k launch",
			cmd:    fmt.Sprintf("k launch -c load-test-http --repoURL %v --set url=https://httpbin.org/get --set duration=2s", srv.URL()),
			golden: base.CompletePath("../testdata", "output/klaunch.txt"),
		},
	}

	runTestActionCmd(t, tests)

}
