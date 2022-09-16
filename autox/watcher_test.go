package autox

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAutoXStart(t *testing.T) {
	addObjectInvocations := 0
	addObject = func(obj interface{}) {
		log.Logger.Debug("Add:", obj)
		addObjectInvocations++
	}

	k8sClient = newFakeKubeClient(cli.New())

	// Start requires some environment variables to be set
	_ = os.Setenv(resourceConfigEnv, "../testdata/autox_inputs/resource_config.example.yaml")
	_ = os.Setenv(chartGroupConfigEnv, "../testdata/autox_inputs/group_config.example.yaml")

	stopCh := make(chan struct{})
	defer close(stopCh)
	_ = Start(stopCh)

	// create object; no track defined
	assert.Equal(t, 0, addObjectInvocations)

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"
	application := "demo"
	version := "v1"
	track := ""

	createdObj, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(namespace, application, version, track),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// give handler time to execute
	assert.Eventually(t, func() bool { return assert.Equal(t, 1, addObjectInvocations) }, 5*time.Second, time.Second)
}
