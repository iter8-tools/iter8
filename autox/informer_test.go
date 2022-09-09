package autox

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewInformer(t *testing.T) {
	Client = *newFakeKubeClient(cli.New())
	w := newIter8Watcher(
		[]schema.GroupVersionResource{{
			Group:    "",
			Version:  "v1",
			Resource: "services",
		}, {
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}},
		[]string{"default", "foo"},
		chartGroupConfig{},
	)
	assert.NotNil(t, w)
}
