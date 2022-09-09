package watcher

import (
	"context"
	"testing"
	"time"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestInformer(t *testing.T) {
	abnapp.Applications.Clear()

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"
	application := "demo"
	version := "v1"
	track := ""

	abnapp.Applications.Clear()

	// define and start watcher
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	w := NewIter8Watcher(
		[]schema.GroupVersionResource{gvr},
		[]string{namespace},
	)
	assert.NotNil(t, w)
	done := make(chan struct{})
	w.Start(done)

	// create object; no track defined
	createdObj, err := k8sclient.Client.Dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(namespace, application, version, track),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// give handler time to execute
	time.Sleep(1 * time.Second)

	// Application object should have been created
	a, err := abnapp.Applications.Get(namespace + "/" + application)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Empty(t, a.Tracks)
	v, _ := a.GetVersion(version, false)
	assert.NotNil(t, v)
	assert.Nil(t, v.Track)
	// assert.Equal(t, track, *v.Track)

	// update object with track
	track = "track"
	(createdObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{})[TRACK_ANNOTATION] = track
	updatedObj, err := k8sclient.Client.Dynamic().
		Resource(gvr).Namespace(namespace).
		Update(
			context.TODO(),
			createdObj,
			metav1.UpdateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, updatedObj)

	time.Sleep(1 * time.Second)

	// now a track is present
	a, err = abnapp.Applications.Get(namespace + "/" + application)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.NotEmpty(t, a.Tracks)
	v, _ = a.GetVersion(version, false)
	assert.NotNil(t, v)
	// assert.Nil(t, v.Track)
	assert.Equal(t, track, *v.Track)

	// delete object --> no track anymore
	err = k8sclient.Client.Dynamic().
		Resource(gvr).Namespace(namespace).
		Delete(
			context.TODO(),
			application,
			metav1.DeleteOptions{},
		)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Application still there but no track
	a, err = abnapp.Applications.Get(namespace + "/" + application)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Empty(t, a.Tracks)
	v, _ = a.GetVersion(version, false)
	assert.NotNil(t, v)
	assert.Nil(t, v.Track)
	// assert.Equal(t, track, *v.Track)

	close(done)
}

func newUnstructuredDeployment(namespace, application, version, track string) *unstructured.Unstructured {
	annotations := map[string]interface{}{
		ITER8_ANNOTATION: "true",
		READY_ANNOTATION: "true",
	}
	if track != "" {
		annotations[TRACK_ANNOTATION] = track
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      application,
				"labels": map[string]interface{}{
					NAME_LABEL:    application,
					VERSION_LABEL: version,
				},
				"annotations": annotations,
			},
			"spec": application,
		},
	}
}
