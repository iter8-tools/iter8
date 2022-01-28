package basecli

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHubGoodFolder(t *testing.T) {
	dir, _ := ioutil.TempDir("", "iter8-test")
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	hubFolder = "load-test"
	// make sure load test folder is present
	err := hubCmd.RunE(nil, nil)
	assert.NoError(t, err)
	_, err = os.Stat(path.Join(dir, hubFolder, "Chart.yaml"))
	assert.False(t, os.IsNotExist(err))
}
