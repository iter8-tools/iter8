package action

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHub(t *testing.T) {
	// fix hOpts
	hOpts := NewHubOpts()
	hOpts.ChartsDir = path.Join(t.TempDir(), chartsFolderName)
	hOpts.GitFolder = "github.com/iter8-tools/iter8.git//charts"

	err := hOpts.LocalRun()
	assert.NoError(t, err)
}
