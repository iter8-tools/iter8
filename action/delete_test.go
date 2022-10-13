package action

import (
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestKubeDelete(t *testing.T) {
	var err error

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartName = base.CompletePath("../", "iter8")
	lOpts.LocalChart = true
	lOpts.Values = []string{"tasks={http}", "http.url=https://iter8.tools", "http.duration=2s"}

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)

	// fix dOpts
	dOpts := NewDeleteOpts(lOpts.KubeDriver)
	err = dOpts.KubeRun()
	assert.NoError(t, err)
}
