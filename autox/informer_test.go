package autox

import (
	"context"
	"testing"
	"time"

	// abnapp "github.com/iter8-tools/iter8/abn/application"
	// "github.com/iter8-tools/iter8/abn/k8sclient"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	trackAnnotation = "iter8.tools/track"
	newLabel        = "app.kubernetes.io/name"
	versionLabel    = "app.kubernetes.io/version"
	readyAnnotation = "iter8.tools/ready"
	iter8Label      = "iter8.tools/abn"
)

// Check to see if add, update, delete handlers from the watcher are properly invoked
func TestWatcher(t *testing.T) {
	addObjectInvocations := 0
	updateObjectInvocations := 0
	deleteObjectInvocations := 0

	// Overwrite original handlers
	addObject = func(obj interface{}) {
		log.Logger.Debug("Add:", obj)
		addObjectInvocations++
	}
	updateObject = func(oldObj, obj interface{}) {
		log.Logger.Debug("Update:", oldObj, obj)
		updateObjectInvocations++
	}
	deleteObject = func(obj interface{}) {
		log.Logger.Debug("Delete:", obj)
		deleteObjectInvocations++
	}

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"
	application := "demo"
	version := "v1"
	track := ""

	// define and start watcher
	k8sClient = newFakeKubeClient(cli.New())
	w := newIter8Watcher(
		[]schema.GroupVersionResource{gvr},
		[]string{namespace},
		chartGroupConfig{},
	)
	assert.NotNil(t, w)
	done := make(chan struct{})
	defer close(done)
	w.start(done)

	// create object; no track defined
	assert.Equal(t, 0, addObjectInvocations)
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

	// update object with track
	assert.Equal(t, 0, updateObjectInvocations)
	track = "track"
	(createdObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{})[trackAnnotation] = track
	updatedObj, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Update(
			context.TODO(),
			createdObj,
			metav1.UpdateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, updatedObj)

	// give handler time to execute
	assert.Eventually(t, func() bool { return assert.Equal(t, 1, updateObjectInvocations) }, 5*time.Second, time.Second)

	// delete object --> no track anymore
	assert.Equal(t, 0, deleteObjectInvocations)
	err = k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Delete(
			context.TODO(),
			application,
			metav1.DeleteOptions{},
		)
	assert.NoError(t, err)

	// give handler time to execute
	assert.Eventually(t, func() bool { return assert.Equal(t, 1, deleteObjectInvocations) }, 5*time.Second, time.Second)
}

func newUnstructuredDeployment(namespace, application, version, track string) *unstructured.Unstructured {
	annotations := map[string]interface{}{
		readyAnnotation: "true",
	}
	if track != "" {
		annotations[trackAnnotation] = track
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      application,
				"labels": map[string]interface{}{
					newLabel:     application,
					versionLabel: version,
					iter8Label:   "true",
				},
				"annotations": annotations,
			},
			"spec": application,
		},
	}
}
