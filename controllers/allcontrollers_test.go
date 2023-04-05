package controllers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStart(t *testing.T) {
	os.Setenv(configEnv, base.CompletePath("../", "testdata/controllers/config.yaml"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := fake.New()
	Start(ctx.Done(), client)

	// create a deployment
	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				iter8WatchLabel: iter8WatchValue,
			},
		},
	}
	_, err := client.Clientset.AppsV1().Deployments("default").Create(ctx, &dep, metav1.CreateOptions{})
	assert.NoError(t, err)

	// create a routemap that changes the replicaset for the deployment
	cm := corev1.ConfigMap{
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
	_, err = client.Clientset.CoreV1().ConfigMaps("default").Create(ctx, &cm, metav1.CreateOptions{})
	assert.NoError(t, err)

	// check if the replicas for the deployment changed
	nd, err := client.Clientset.AppsV1().Deployments("default").Get(ctx, dep.Name, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Eventually(t, func() bool {
		return assert.Equal(t, 2, nd.Spec.Replicas)
	}, time.Second*2, time.Millisecond*100)

	// check if finalizer has been added
	assert.Eventually(t, func() bool {
		return assert.Contains(t, nd.ObjectMeta.Finalizers, iter8FinalizerStr)
	}, time.Second*2, time.Millisecond*100)

}
