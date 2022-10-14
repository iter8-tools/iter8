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
	// autoX watcher will call on installHelmRelease
	installHelmReleaseInvocations := 0
	installHelmRelease = func(releaseName string, chartGroupName string, chart chart, namespace string) error {
		installHelmReleaseInvocations++
		return nil
	}

	opts := NewOpts(newFakeKubeClient(cli.New()))

	// Start requires some environment variables to be set
	_ = os.Setenv(chartGroupConfigEnv, "../testdata/autox_inputs/group_config.example.yaml")

	stopCh := make(chan struct{})
	defer close(stopCh)
	_ = opts.Start(stopCh)

	// create object; no track defined
	assert.Equal(t, 0, installHelmReleaseInvocations)

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"
	application := "demo"
	version := "v1"
	track := ""

	createdObj, err := opts.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				application,
				version,
				track,
				map[string]string{
					// add the autoXLabel, which will allow installHelmRelease to trigger
					autoXLabel: "myApp",
				},
			), metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// give handler time to execute
	// creating an object will installHelmRelease for each chart in the chart group
	// in this case, there are 2 charts
	// once for autox-myApp-name1-XXXXX and autox-myApp-name2-XXXXX
	assert.Eventually(t, func() bool { return assert.Equal(t, 2, installHelmReleaseInvocations) }, 5*time.Second, time.Second)
}
