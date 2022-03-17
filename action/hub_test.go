package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHub(t *testing.T) {
	srv := SetupWithRepo(t)

	// fix hOpts
	hOpts := NewHubOpts()
	hOpts.DestDir = t.TempDir()
	hOpts.ChartName = "load-test-http"
	hOpts.RepoURL = srv.URL()

	err := hOpts.LocalRun()
	assert.NoError(t, err)
}
