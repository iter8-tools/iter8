package action

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHub(t *testing.T) {
	// fix hOpts
	hOpts := NewHubOpts()
	_ = os.Chdir(t.TempDir())
	hOpts.RemoteFolderURL = "github.com/iter8-tools/iter8.git//charts"

	err := hOpts.LocalRun()
	assert.NoError(t, err)
}
