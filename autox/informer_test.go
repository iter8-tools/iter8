package autox

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"

// 	"helm.sh/helm/v3/pkg/cli"
// 	"k8s.io/apimachinery/pkg/runtime/schema"
// )

// func TestNewInformer(t *testing.T) {
// 	Client = *newFakeKubeClient(cli.New())
// 	w := newIter8Watcher(
// 		[]schema.GroupVersionResource{{
// 			Group:    "",
// 			Version:  "v1",
// 			Resource: "services",
// 		}, {
// 			Group:    "apps",
// 			Version:  "v1",
// 			Resource: "deployments",
// 		}},
// 		[]string{"default", "foo"},
// 		chartGroupConfig{},
// 	)
// 	assert.NotNil(t, w)
// }

import (
	"context"
	"testing"
	"time"

	// abnapp "github.com/iter8-tools/iter8/abn/application"
	// "github.com/iter8-tools/iter8/abn/k8sclient"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TRACK_ANNOTATION = "iter8.tools/track"
var NAME_LABEL = "app.kubernetes.io/name"
var VERSION_LABEL = "app.kubernetes.io/version"
var READY_ANNOTATION = "iter8.tools/ready"
var ITER8_ANNOTATION = "iter8.tools/abn"
var ITER8_LABEL = "iter8.tools/abn"

func TestInformer(t *testing.T) {
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
	Client = *newFakeKubeClient(cli.New())
	w := newIter8Watcher(
		[]schema.GroupVersionResource{gvr},
		[]string{namespace},
		chartGroupConfig{},
	)
	assert.NotNil(t, w)
	done := make(chan struct{})
	w.start(done)

	// create object; no track defined
	assert.Equal(t, 0, addObjectInvocations)
	createdObj, err := Client.Dynamic().
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
	assert.Equal(t, 1, addObjectInvocations)

	// update object with track
	assert.Equal(t, 0, updateObjectInvocations)
	track = "track"
	(createdObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{})[TRACK_ANNOTATION] = track
	updatedObj, err := Client.Dynamic().
		Resource(gvr).Namespace(namespace).
		Update(
			context.TODO(),
			createdObj,
			metav1.UpdateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, updatedObj)

	// give handler time to execute
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, updateObjectInvocations)

	// delete object --> no track anymore
	assert.Equal(t, 0, deleteObjectInvocations)
	err = Client.Dynamic().
		Resource(gvr).Namespace(namespace).
		Delete(
			context.TODO(),
			application,
			metav1.DeleteOptions{},
		)
	assert.NoError(t, err)

	// give handler time to execute
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, deleteObjectInvocations)

	close(done)
}

func newUnstructuredDeployment(namespace, application, version, track string) *unstructured.Unstructured {
	annotations := map[string]interface{}{
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
					ITER8_LABEL:   "true",
				},
				"annotations": annotations,
			},
			"spec": application,
		},
	}
}
