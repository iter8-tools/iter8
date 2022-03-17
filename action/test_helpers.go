package action

import (
	"testing"

	"github.com/iter8-tools/iter8/base"
	"helm.sh/helm/v3/pkg/repo/repotest"
)

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
