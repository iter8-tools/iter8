package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHub(t *testing.T) {
	// fix hOpts
	hOpts := NewHubOpts()
	hOpts.ChartsParentDir = t.TempDir()
	hOpts.GitFolder = "github.com/iter8-tools/iter8.git//charts"

	err := hOpts.LocalRun()
	assert.NoError(t, err)
}
