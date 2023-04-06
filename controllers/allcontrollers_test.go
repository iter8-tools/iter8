package controllers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestStart(t *testing.T) {
	// set pod name
	_ = os.Setenv(PodNameEnvVariable, "pod-0")
	// set pod namespace
	_ = os.Setenv(PodNamespaceEnvVariable, "default")
	// set config file
	_ = os.Setenv(ConfigEnv, base.CompletePath("../", "testdata/controllers/config.yaml"))

	// make a routemap that manages replicas for deployment
	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				iter8ManagedByLabel: iter8ManagedByValue,
				iter8KindLabel:      iter8KindRoutemapValue,
				iter8VersionLabel:   iter8VersionValue,
			},
		},
		Immutable: base.BoolPointer(true),
		Data: map[string]string{
			"strSpec": `
variants:
- resources:
  - gvrShort: deploy
    name: test
    namespace: default
routingTemplates:
  test:
    gvrShort: deploy
    template: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: test
        namespace: default
      spec:
        replicas: 2
`,
		},
		BinaryData: map[string][]byte{},
	}

	// make a deployment
	depu := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "default",
				"labels": map[string]interface{}{
					iter8WatchLabel: iter8WatchValue,
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := fake.New([]runtime.Object{&cm}, []runtime.Object{depu})
	err := Start(ctx.Done(), client)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		// check if the replicas for the deployment changed
		nd, _ := client.FakeDynamicClient.Resource(schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}).Namespace("default").Get(context.Background(), "test", metav1.GetOptions{})

		log.Logger.Debug("uns: ", nd)

		if nd != nil {
			rep, a, b := unstructured.NestedInt64(nd.UnstructuredContent(), "spec", "replicas")
			if !a || b != nil {
				return false
			}
			return assert.Equal(t, int64(2), rep)
		}

		return false
	}, time.Second*2, time.Millisecond*100)

	// check if finalizer has been added
	assert.Eventually(t, func() bool {
		nd, _ := client.FakeDynamicClient.Resource(schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}).Namespace("default").Get(context.Background(), "test", metav1.GetOptions{})

		log.Logger.Debug("uns: ", nd)

		if nd != nil {
			finalizers, a, b := unstructured.NestedStringSlice(nd.UnstructuredContent(), "metadata", "finalizers")
			if !a || b != nil {
				return false
			}
			return assert.Contains(t, finalizers, iter8FinalizerStr)
		}

		return false
	}, time.Second*2, time.Millisecond*100)
}
