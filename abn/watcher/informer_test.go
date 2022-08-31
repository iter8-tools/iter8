package watcher

import (
	"testing"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewInformer(t *testing.T) {
	kClient := k8sclient.NewFakeKubeClient(cli.New())
	w := NewIter8Watcher(
		kClient,
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
	)
	assert.NotNil(t, w)
}
