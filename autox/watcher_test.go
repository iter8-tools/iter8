package autox

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestStart(t *testing.T) {
	// Start() requires some environment variables to be set
	_ = os.Setenv(configEnv, "../testdata/autox_inputs/config.example.yaml")

	stopCh := make(chan struct{})
	defer close(stopCh)
	_ = Start(stopCh, newFakeKubeClient(cli.New()))

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"
	releaseSpecName := "myApp"
	version := "v1"
	track := ""

	// create releaseSpec secret
	releaseGroupSpecSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: argocd,
			Labels: map[string]string{
				"iter8.tools/autox-group": releaseSpecName,
			},
		},
	}
	_, err := k8sClient.clientset.CoreV1().Secrets(argocd).Create(context.Background(), &releaseGroupSpecSecret, metav1.CreateOptions{})
	assert.NoError(t, err)

	createdObj, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				releaseSpecName,
				version,
				track,
				map[string]string{
					// autoXLabel: "true", // add the autoXLabel, which will allow applyApplication() to trigger
				},
			), metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// 2 applications
	// one for each release spec in the config
	// autox-myapp-name1 and autox-myapp-name2
	assert.Eventually(t, func() bool {
		list, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).List(context.Background(), metav1.ListOptions{})
		return assert.Equal(t, len(list.Items), 2)
	}, 5*time.Second, time.Second)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		c   config
		err string
	}{
		{
			config{
				Specs: map[string]releaseGroupSpec{
					"test": {},
				},
			},
			"trigger in spec group \"test\" does not have a name",
		},
		{
			config{
				Specs: map[string]releaseGroupSpec{
					"test": {
						Trigger: trigger{
							Name: "test",
						},
					},
				},
			},
			"trigger in spec group \"test\" does not have a namespace",
		},
		{
			config{
				Specs: map[string]releaseGroupSpec{
					"test": {
						Trigger: trigger{
							Name:      "test",
							Namespace: "default",
						},
					},
				},
			},
			"trigger in spec group \"test\" does not have a version",
		},
		{
			config{
				Specs: map[string]releaseGroupSpec{
					"test": {
						Trigger: trigger{
							Name:      "test",
							Namespace: "default",
							Version:   "v1",
						},
					},
				},
			},
			"trigger in spec group \"test\" does not have a resource",
		},
		{
			config{
				Specs: map[string]releaseGroupSpec{
					"test": {
						Trigger: trigger{
							Name:      "test",
							Namespace: "default",
							Version:   "v1",
							Resource:  "deployments",
						},
					},
					"test2": {
						Trigger: trigger{
							Name:      "test",
							Namespace: "default",
							Version:   "v1",
							Resource:  "deployments",
						},
					},
				},
			},
			"multiple release specs with the same trigger: name: \"test\", namespace: \"default\", group: \"\", version: \"v1\", resource: \"deployments\",",
		},
	}

	for _, e := range tests {
		err := validateConfig(e.c)
		fmt.Println(err)
		assert.EqualError(t, err, e.err)
	}
}
