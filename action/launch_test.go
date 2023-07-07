package action

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestKubeLaunch(t *testing.T) {
	var err error
	_ = os.Chdir(t.TempDir())

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.Values = []string{"tasks={http}", "http.url=https://httpbin.org/get", "http.duration=2s", "runner=job"}

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)
}

func TestKubeLaunchLocalChart(t *testing.T) {
	var err error
	_ = os.Chdir(t.TempDir())

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartName = base.CompletePath("../charts", "iter8")
	lOpts.LocalChart = true
	lOpts.Values = []string{"tasks={http}", "http.url=https://httpbin.org/get", "http.duration=2s", "runner=job"}

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)
}
