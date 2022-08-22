package watcher

import (
	"testing"

	k8sdriver "github.com/iter8-tools/iter8/base/k8sdriver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewInformer(t *testing.T) {
	kd := k8sdriver.NewFakeKubeDriver(cli.New())
	w := NewIter8Watcher(
		kd,
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
