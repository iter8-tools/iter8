package watcher

import (
	"testing"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewInformer(t *testing.T) {
	k8sclient.Client = *k8sclient.NewFakeKubeClient()
	w := NewIter8Watcher(
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
