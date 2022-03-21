package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
)

func TestHub(t *testing.T) {
	srv := id.SetupWithRepo(t)
	os.Chdir(t.TempDir())

	tests := []cmdTestCase{
		// hub
		{
			name:   "basic hub",
			cmd:    fmt.Sprintf("hub -c load-test-http --repoURL %v", srv.URL()),
			golden: base.CompletePath("../testdata", "output/hub.txt"),
		},
		// hub, destDir
		{
			name:   "hub with destDir",
			cmd:    fmt.Sprintf("hub -c load-test-http --destDir %v --repoURL %v", t.TempDir(), srv.URL()),
			golden: base.CompletePath("../testdata", "output/hub-with-destdir.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
