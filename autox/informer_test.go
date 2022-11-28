package autox

import (

	// abnapp "github.com/iter8-tools/iter8/abn/application"
	// "github.com/iter8-tools/iter8/abn/k8sclient"

	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Check to see if add, update, delete handlers from the watcher are properly invoked
// after the watcher is created using newIter8Watcher()
func TestNewIter8Watcher(t *testing.T) {
	// autoX needs the config
	autoXConfig := readConfig("../testdata/autox_inputs/config.example.yaml")

	// autoX handler will call on installHelmRelease and deleteHelmRelease
	installHelmReleaseInvocations := 0
	installHelmRelease = func(releaseName string, specGroupName string, releaseSpec releaseSpec, namespace string, additionalValues map[string]string) error {
		installHelmReleaseInvocations++
		return nil
	}

	deleteHelmReleaseInvocations := 0
	deleteHelmRelease = func(releaseName string, specGroupName string, namespace string) error {
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

	// create object without autoXLabel
	// this should not trigger any installHelmRelease (or deleteHelmRelease)
	assert.Equal(t, 0, installHelmReleaseInvocations)
	assert.Equal(t, 0, deleteHelmReleaseInvocations)
	createdObjNoAutoXLabel, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				"demo",
				version,
				track,
				map[string]string{},
			),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObjNoAutoXLabel)

	// give handler time to execute
	// the invocations should not change as the created object has no autoXLabel
	assert.Eventually(t, func() bool { return assert.Equal(t, 0, installHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 0, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// create object with autoXLabel
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
					// which will allow installHelmRelease and deleteHelmRelease to trigger
					autoXLabel: "myApp",
				},
			),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// give handler time to execute
	// creating an object will installHelmRelease for each spec in the spec group
	// in this case, there are 2 specs
	// once for autox-myApp-name1-XXXXX and autox-myApp-name2-XXXXX
	assert.Eventually(t, func() bool { return assert.Equal(t, 2, installHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 2, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// update annotations
	// this should not trigger deleteHelmRelease/installHelmRelease
	// these functions should only be triggered when a (pruned) label is changed
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
	// the invocations should not change as a (pruned) label was not changed
	assert.Eventually(t, func() bool { return assert.Equal(t, 4, installHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 4, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// update (pruned) labels
	// this should trigger deleteHelmRelease/installHelmRelease
	// change versionLabel, which is a pruned label
	(createdObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[versionLabel] = "v2"
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
	// updating (pruned) labels will trigger both deleteHelmRelease and installHelmRelease for each spec in the spec group
	// in this case, there are 2 specs
	assert.Eventually(t, func() bool { return assert.Equal(t, 6, installHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 6, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)

	// delete object
	err = k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Delete(
			context.TODO(),
			application,
			metav1.DeleteOptions{},
		)
	assert.NoError(t, err)

	// give handler time to execute
	// only deleteHelmRelease should trigger, once for each spec in the spec group
	// in this case, there are 2 specs
	assert.Eventually(t, func() bool { return assert.Equal(t, 6, installHelmReleaseInvocations) }, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool { return assert.Equal(t, 8, deleteHelmReleaseInvocations) }, 5*time.Second, time.Second)
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
