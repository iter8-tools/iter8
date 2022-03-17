package action

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestLocalLaunch(t *testing.T) {
	SetupWithMock(t)
	srv := SetupWithRepo(t)
	os.Chdir(t.TempDir())

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartName = "load-test-http"
	lOpts.Values = []string{"url=https://httpbin.org/get", "duration=2s"}
	lOpts.RepoURL = srv.URL()

	err := lOpts.LocalRun()
	assert.NoError(t, err)
}

func TestKubeLaunch(t *testing.T) {
	srv := SetupWithRepo(t)

	var err error

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartName = "load-test-http"
	lOpts.DestDir = t.TempDir()
	lOpts.Values = []string{"url=https://iter8.tools", "duration=2s"}
	lOpts.RepoURL = srv.URL()

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)
}
