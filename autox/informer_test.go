package autox

import (
	// abnapp "github.com/iter8-tools/iter8/abn/application"
	// "github.com/iter8-tools/iter8/abn/k8sclient"

	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestShouldCreateApplication tests the function shouldCreateApplication(), which determines if an application should
// created/updated based on whether or not there is a preexisting one
func TestShouldCreateApplication(t *testing.T) {
	// 1) nothing in cluster
	// therefore, return true (no concern for preexisting application)
	k8sClient = newFakeKubeClient(cli.New())
	assert.True(t, shouldCreateApplication(map[string]interface{}{}, "test"))

	// 2) existing application, new application has the same values
	// therefore, return false (not necessary to recreate application)
	values := applicationValues{
		Name:      "test",
		Namespace: "default",
		Chart: releaseSpec{
			Values: map[string]interface{}{},
		},
	}

	// simulating additional values
	values.Chart.Values["hello"] = "world"

	// execute application template
	uApp, err := executeApplicationTemplate(tplStr, values)
	assert.NoError(t, err)
	_, err = k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Create(context.Background(), uApp, metav1.CreateOptions{})
	assert.NoError(t, err)

	// same values (values.Chart.Values)
	// therefore, return false
	assert.False(t, shouldCreateApplication(values.Chart.Values, "test"))

	// 3) existing application, new application has different values
	// therefore, return true (old application can be replaced with new one)

	// different values
	// therefore, return true
	assert.True(t, shouldCreateApplication(map[string]interface{}{"something": "different"}, "test"))

	// 4) existing application but application is not managed by Iter8
	// therefore return false (Iter8 does not have permission to replace the old application)

	// setting managed by to something other than Iter8
	uApp.SetLabels(map[string]string{
		managedByLabel: "abc",
	})

	_, err = k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Update(context.Background(), uApp, metav1.UpdateOptions{})
	assert.NoError(t, err)

	assert.False(t, shouldCreateApplication(map[string]interface{}{"something": "different"}, "test"))
}

// TestApplyApplication tests the function applyApplication(), which applys Argo CD applications
func TestApplyApplication(t *testing.T) {
	k8sClient = newFakeKubeClient(cli.New())

	releaseGroupSpecName := "testReleaseGroupSpecName"
	releaseSpecName := "testReleaseSpecName"
	applicationName := fmt.Sprintf("autox-%s-%s", releaseGroupSpecName, releaseSpecName)
	spec := releaseSpec{
		Name:   applicationName,
		Values: map[string]interface{}{},
	}
	additionalValues := map[string]interface{}{}

	// 1) no release group spec secret
	// therefore, fail
	assert.Error(t, applyApplication(applicationName, releaseGroupSpecName, spec, "default", additionalValues))

	// 2) create application with no conflicts
	// create release group spec secret
	// therefore, no fail
	releaseGroupSpecSecret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: argocd,
			Labels: map[string]string{
				"iter8.tools/autox-group": releaseGroupSpecName,
			},
		},
	}

	_, err := k8sClient.clientset.CoreV1().Secrets(argocd).Create(context.Background(), &releaseGroupSpecSecret, metav1.CreateOptions{})
	assert.NoError(t, err)

	// ensure application does not exist
	_, err = k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Get(context.Background(), applicationName, metav1.GetOptions{})
	assert.Error(t, err)

	assert.NoError(t, applyApplication(applicationName, releaseGroupSpecName, spec, "default", additionalValues))

	// ensure application exists
	_, err = k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Get(context.Background(), applicationName, metav1.GetOptions{})
	assert.NoError(t, err)

	// 3) create application with conflicts
	// fallback is to do nothing
	// therefore, no fail
	assert.NoError(t, applyApplication(applicationName, releaseGroupSpecName, spec, "default", additionalValues))
}

// TestDeleteApplication tests the function deleteApplication(), which deletes Argo CD applications
func TestDeleteApplication(t *testing.T) {
	k8sClient = newFakeKubeClient(cli.New())

	releaseGroupSpecName := "testReleaseGroupSpecName"
	releaseSpecName := "testReleaseSpecName"
	applicationName := fmt.Sprintf("autox-%s-%s", releaseGroupSpecName, releaseSpecName)

	// 1) no application
	// therefore, fail
	assert.Error(t, deleteApplication(applicationName))

	// 2) delete existing application
	// therefore, no fail

	// create application
	values := applicationValues{
		Name: applicationName,
		Chart: releaseSpec{
			Name:   applicationName,
			Values: map[string]interface{}{},
		},
	}
	uApp, err := executeApplicationTemplate(tplStr, values)
	assert.NoError(t, err)
	_, err = k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Create(context.TODO(), uApp, metav1.CreateOptions{})
	assert.NoError(t, err)

	// ensure there is an application
	_, err = k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Get(context.Background(), applicationName, metav1.GetOptions{})
	assert.NoError(t, err)

	assert.NoError(t, deleteApplication(applicationName))

	// ensure there is no application anymore
	_, err = k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Get(context.Background(), applicationName, metav1.GetOptions{})
	assert.Error(t, err)
}

// Check to see if add, update, delete handlers from the watcher are properly invoked
// after the watcher is created using newIter8Watcher()
func TestNewIter8Watcher(t *testing.T) {
	// autoX needs the config
	autoXConfig := readConfig("../testdata/autox_inputs/config.example.yaml")

	namespace := "default"
	releaseSpecName := "myApp"
	version := "v1"
	track := ""
	appName1 := "autox-myApp-name1"
	appName2 := "autox-myApp-name2"

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	// define and start watcher
	k8sClient = newFakeKubeClient(cli.New())

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

	w := newIter8Watcher(autoXConfig)
	assert.NotNil(t, w)
	done := make(chan struct{})
	defer close(done)
	w.start(done)

	// 1) create object with random name and no version label
	// no application should be created
	objRandNameNoAutoXLabel, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				"rand", // random name
				"",     // no version label
				track,
				map[string]string{},
			),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, objRandNameNoAutoXLabel)

	// no applications
	assert.Eventually(t, func() bool {
		list, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).List(context.Background(), metav1.ListOptions{})
		return assert.Equal(t, len(list.Items), 0)
	}, 5*time.Second, time.Second)

	// 2) create object with random name and version label
	// no application should be created
	objRandNameAutoXLabel, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				"rand2", // random name
				version, // version label
				track,
				map[string]string{},
			),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, objRandNameAutoXLabel)

	// no applications
	assert.Eventually(t, func() bool {
		list, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).List(context.Background(), metav1.ListOptions{})
		return assert.Equal(t, len(list.Items), 0)
	}, 5*time.Second, time.Second)

	// 3) create object with trigger name and no version label
	// no application should be created
	objNoAutoXLabel, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				releaseSpecName, // trigger name
				"",              // no version label
				track,
				map[string]string{}),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, objNoAutoXLabel)

	// no applications
	assert.Eventually(t, func() bool {
		list, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).List(context.Background(), metav1.ListOptions{})
		return assert.Equal(t, len(list.Items), 0)
	}, 5*time.Second, time.Second)

	// delete the object so we can recreate it with autoX label
	err = k8sClient.dynamic().Resource(gvr).Namespace(namespace).Delete(context.TODO(), releaseSpecName, metav1.DeleteOptions{})
	assert.NoError(t, err)

	// 4) create object with trigger name with version label
	// 2 applications should be created
	// one for each release spec in the config
	// autox-myapp-name1 and autox-myapp-name2
	createdObj, err := k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(
				namespace,
				releaseSpecName, // trigger name
				version,         // version label
				track,
				map[string]string{},
			),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	// 2 applications
	assert.Eventually(t, func() bool {
		list, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).List(context.Background(), metav1.ListOptions{})
		return assert.Equal(t, len(list.Items), 2)
	}, 5*time.Second, time.Second)

	// check applications by name
	assert.Eventually(t, func() bool {
		app, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Get(context.Background(), appName1, metav1.GetOptions{})
		return assert.NotNil(t, app)
	}, 5*time.Second, time.Second)
	assert.Eventually(t, func() bool {
		app, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Get(context.Background(), appName2, metav1.GetOptions{})
		return assert.NotNil(t, app)
	}, 5*time.Second, time.Second)

	// 5) delete version label
	// all applications deleted
	(createdObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[versionLabel] = nil
	_, err = k8sClient.dynamic().
		Resource(gvr).Namespace(namespace).
		Update(
			context.TODO(),
			createdObj,
			metav1.UpdateOptions{},
		)
	assert.NoError(t, err)

	// 0 applications
	assert.Eventually(t, func() bool {
		list, _ := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).List(context.Background(), metav1.ListOptions{})
		return assert.Equal(t, len(list.Items), 0)
	}, 5*time.Second, time.Second)
}

func newUnstructuredDeployment(namespace, application, version, track string, additionalLabels map[string]string) *unstructured.Unstructured {
	annotations := map[string]interface{}{
		"iter8.tools/ready": "true",
	}
	if track != "" {
		annotations[trackLabel] = track
	}

	labels := map[string]interface{}{
		nameLabel:           application,
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
