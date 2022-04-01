package action

import (
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestLocalLaunch(t *testing.T) {
	base.SetupWithMock(t)

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartsParentDir = base.CompletePath("../", "")
	lOpts.ChartName = "load-test-http"
	lOpts.NoDownload = true
	lOpts.Values = []string{"url=https://httpbin.org/get", "duration=2s"}
	lOpts.RunDir = t.TempDir()

	err := lOpts.LocalRun()
	assert.NoError(t, err)
}

func TestKubeLaunch(t *testing.T) {
	var err error

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartsParentDir = base.CompletePath("../", "")
	lOpts.ChartName = "load-test-http"
	lOpts.NoDownload = true
	lOpts.Values = []string{"url=https://httpbin.org/get", "duration=2s"}
	lOpts.RunDir = t.TempDir()

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)
}
