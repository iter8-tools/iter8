package basecli

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHubGoodFolder(t *testing.T) {
	chartName = "load-test-http"
	// make sure load test folder is present
	err := hubCmd.RunE(nil, nil)
	assert.NoError(t, err)
	_, err = os.Stat(path.Join(iter8TempDir, chartName, "Chart.yaml"))
	assert.False(t, os.IsNotExist(err))
}
