package action

import (
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/jarcoal/httpmock"
	"helm.sh/helm/v3/pkg/repo/repotest"
)

func SetupWithMock(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))
	t.Cleanup(httpmock.DeactivateAndReset)
}

func SetupWithRepo(t *testing.T) *repotest.Server {
	srv, err := repotest.NewTempServerWithCleanup(t, base.CompletePath("../", "testdata/charts/*.tgz*"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(srv.Stop)

	if err := srv.LinkIndices(); err != nil {
		t.Fatal(err)
	}
	return srv
}
