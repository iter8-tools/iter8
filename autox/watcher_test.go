package autox

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestStart(t *testing.T) {
	// autoX watcher will call on applyHelmRelease
	applyHelmReleaseInvocations := 0
	applyApplicationObject = func(releaseName string, group string, releaseSpec releaseSpec, namespace string, additionalValues map[string]interface{}) error {
		applyHelmReleaseInvocations++
		return nil
	}

	// Start requires some environment variables to be set
	_ = os.Setenv(configEnv, "../testdata/autox_inputs/config.example.yaml")

	stopCh := make(chan struct{})
	defer close(stopCh)
	_ = Start(stopCh, newFakeKubeClient(cli.New()))

	// create object; no track defined
	assert.Equal(t, 0, applyHelmReleaseInvocations)

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"
	application := "myApp"
	version := "v1"
	track := ""

	createdObj, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				application,
				version,
				track,
				map[string]string{
					// add the autoXLabel, which will allow applyHelmRelease to trigger
					autoXLabel: "myApp",
				},
			), metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// give handler time to execute
	// creating an object will applyHelmRelease for each spec in the spec group
	// in this case, there are 2 specs
	// once for autox-myApp-name1-XXXXX and autox-myApp-name2-XXXXX
	assert.Eventually(t, func() bool { return assert.Equal(t, 2, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
}
