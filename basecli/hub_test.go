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
	// Todo: fix location below
	os.Setenv("ITER8HUB", "github.com/sriumcp/iter8.git?ref=v0.8//hub/")
	hubFolder = "load-test"
	// make sure load test folder is present
	err := hubCmd.RunE(nil, nil)
	assert.NoError(t, err)
	_, err = os.Stat(path.Join(dir, hubFolder, "experiment.yaml"))
	assert.False(t, os.IsNotExist(err))

	hubFolder = "random-loc"
	// make sure proper error is generated
	err = hubCmd.RunE(nil, nil)
	assert.Error(t, err)
}
