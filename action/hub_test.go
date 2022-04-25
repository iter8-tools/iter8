package action

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHub(t *testing.T) {
	// fix hOpts
	hOpts := NewHubOpts()
	os.Chdir(t.TempDir())
	hOpts.Folder = "github.com/iter8-tools/iter8.git//charts"

	err := hOpts.LocalRun()
	assert.NoError(t, err)
}
