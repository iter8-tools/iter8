package action

import (
	"testing"

	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestLocalLaunch(t *testing.T) {
	SetupWithMock(t)
	srv := SetupWithRepo(t)
	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartName = "load-test-http"
	lOpts.DestDir = t.TempDir()
	lOpts.Values = []string{"url=https://httpbin.org/get", "duration=2s"}
	lOpts.RepoURL = srv.URL()

	err := lOpts.LocalRun()
	assert.NoError(t, err)
}

func TestKubeLaunch(t *testing.T) {
	SetupWithMock(t)
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
	assert.NoError(t, err)
}
