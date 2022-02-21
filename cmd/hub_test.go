package cmd

import (
	"os"
	"path"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/repo/repotest"
)

func TestHubGoodFolder(t *testing.T) {
	srv, err := repotest.NewTempServerWithCleanup(t, base.CompletePath("../", "testdata/charts/*.tgz*"))
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()
	if err = srv.CreateIndex(); err != nil {
		t.Fatal(err)
	}
	if err = srv.LinkIndices(); err != nil {
		t.Fatal(err)
	}

	repoURL = srv.URL()
	chartName = "load-test-http"
	chartVersionConstraint = "0.1.0"

	defer cleanChartArtifacts(destDir, chartName)

	// make sure load test folder is present
	err = hubCmd.RunE(nil, nil)
	assert.NoError(t, err)
	_, err = os.Stat(path.Join(destDir, chartName, "Chart.yaml"))
	assert.False(t, os.IsNotExist(err))
}
