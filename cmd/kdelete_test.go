package cmd

import (
	"fmt"
	"testing"

	id "github.com/iter8-tools/iter8/driver"
)

func TestKDelete(t *testing.T) {
	srv := id.SetupWithRepo(t)

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
			name: "delete",
			cmd:  "k delete",
		},
	}

	runTestActionCmd(t, tests)
}
