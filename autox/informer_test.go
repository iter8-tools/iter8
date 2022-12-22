package autox

import (

	// abnapp "github.com/iter8-tools/iter8/abn/application"
	// "github.com/iter8-tools/iter8/abn/k8sclient"

	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Check to see if add, update, delete handlers from the watcher are properly invoked
// after the watcher is created using newIter8Watcher()
func TestNewIter8Watcher(t *testing.T) {
	// autoX needs the config
	autoXConfig := readConfig("../testdata/autox_inputs/config.example.yaml")

	// autoX handler will call on applyHelmRelease and deleteHelmRelease
	applyHelmReleaseInvocations := 0
	applyApplicationObject = func(releaseName string, specGroupName string, releaseSpec releaseSpec, namespace string, additionalValues map[string]interface{}) error {
		applyHelmReleaseInvocations++
		return nil
	}

	deleteHelmReleaseInvocations := 0
	deleteApplicationObject = func(releaseName string, specGroupName string) error {
		deleteHelmReleaseInvocations++
		return nil
	}

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"
	application := "myApp"
	version := "v1"
	track := ""

	// define and start watcher
	k8sClient = newFakeKubeClient(cli.New())

	w := newIter8Watcher(autoXConfig)
	assert.NotNil(t, w)
	done := make(chan struct{})
	defer close(done)
	w.start(done)

	// create object with random name and no autoX label
	// this should not trigger any applyHelmRelease or deleteHelmRelease
	assert.Equal(t, 0, applyHelmReleaseInvocations)
	assert.Equal(t, 0, deleteHelmReleaseInvocations)
	objRandNameNoAutoXLabel, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				"rand", // random name
				version,
				track,
				map[string]string{}, // no autoX label
			),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, objRandNameNoAutoXLabel)

	// give handler time to execute
	// this should trigger applyHelmRelease or deleteHelmRelease because the object does not have the trigger name
	assert.Eventually(t, func() bool { return assert.Equal(t, 0, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 0, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// create object with random name and autoX label
	// this should not trigger any applyHelmRelease or deleteHelmRelease
	objRandNameAutoXLabel, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				"rand2", // random name
				version,
				track,
				map[string]string{
					autoXLabel: "myApp", // autoX label
				}),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, objRandNameAutoXLabel)

	// give handler time to execute
	// this should trigger applyHelmRelease or deleteHelmRelease because the object does not have the trigger name
	assert.Eventually(t, func() bool { return assert.Equal(t, 0, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 0, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// create object with trigger name and no autoX label
	// this should trigger deleteHelmRelease
	objNoAutoXLabel, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				application, // trigger name
				version,
				track,
				map[string]string{}),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, objNoAutoXLabel)

	// give handler time to execute
	// this should trigger deleteHelmRelease because the object does not have the autoX label
	// trigger twice for each release spec
	assert.Eventually(t, func() bool { return assert.Equal(t, 0, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 2, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// delete the object so we can recreate it with autoX label
	err = k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).Delete(context.TODO(), application, metav1.DeleteOptions{})
	assert.NoError(t, err)

	assert.Eventually(t, func() bool { return assert.Equal(t, 0, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 4, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// create object with trigger name and autoX label
	createdObj, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				application, // trigger name
				version,
				track,
				map[string]string{
					autoXLabel: "myApp", // autoX label
				},
			),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// give handler time to execute
	// this should trigger applyHelmRelease but not deleteHelmRelease
	// trigger twice for each release spec
	assert.Eventually(t, func() bool { return assert.Equal(t, 2, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 4, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// arbitrary update (but not the autoX label)
	track = "track"
	(createdObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{})[trackLabel] = track
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
	// this should trigger applyHelmRelease
	// trigger twice for each release spec
	assert.Eventually(t, func() bool { return assert.Equal(t, 4, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 4, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// delete autoX label
	(createdObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[autoXLabel] = nil
	updatedObj, err = k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Update(
			context.TODO(),
			createdObj,
			metav1.UpdateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, updatedObj)

	// give handler time to execute
	// this should trigger deleteHelmRelease
	// trigger twice for each release spec
	assert.Eventually(t, func() bool { return assert.Equal(t, 4, applyHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 6, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)
}

func newUnstructuredDeployment(namespace, application, version, track string, additionalLabels map[string]string) *unstructured.Unstructured {
	annotations := map[string]interface{}{
		"iter8.tools/ready": "true",
	}
	if track != "" {
		annotations[trackLabel] = track
	}

	labels := map[string]interface{}{
		appLabel:            application,
		versionLabel:        version,
		"iter8.tools/ready": "true",
	}

	// add additionalLabels to labels
	if len(additionalLabels) > 0 {
		for labelName, labelValue := range additionalLabels {
			labels[labelName] = labelValue
		}
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace":   namespace,
				"name":        application,
				"labels":      labels,
				"annotations": annotations,
			},
			"spec": application,
		},
	}
}
