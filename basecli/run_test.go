package basecli

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	dir, _ := ioutil.TempDir("", "iter8-test")
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	// Todo: fix location below
	os.Setenv("ITER8HUB", "github.com/sriumcp/iter8.git?ref=v0.8//hub/")
	hubFolder = "load-test"
	// make sure load test folder is present
	err := hubCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// get into experiment folder and run
	os.Chdir(path.Join(dir, hubFolder))
	err = runCmd.RunE(nil, nil)
	assert.NoError(t, err)
}
