package watcher

import (
	"context"
	"fmt"
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

// If sufficient objects created to match config, then Applications should contain suitable entry
// Probably could test various combinations
func TestAdd(t *testing.T) {
	var svcConfigFile = "../../testdata/abninputs/config.yaml"
	var namespace = "default"
	var application = "backend"
	var track = application
	var version = "v1"

	// set up - initialize channel
	done := make(chan struct{})
	defer close(done)

	// set up - Applications and informers
	setup(t, svcConfigFile, done)

	// create objects in cluster
	createObject(t, newUnstructuredDeployment(namespace, application, version))
	createObject(t, newUnstructuredService(namespace, application, version))

	// creation of these objects should trigger handler which will add application to Applications map
	assert.Eventually(t, func() bool {
		return assertApplicationExists(t, namespace, application, []string{track}, []string{version})
	}, 10*time.Second, 100*time.Millisecond)
}

// update version associated with track
func TestUpdate(t *testing.T) {
	var svcConfigFile = "../../testdata/abninputs/config.yaml"
	var namespace = "default"
	var application = "backend"
	var track = application
	var version = "v1"
	var version2 = "v2"

	// set up - initialize channel
	done := make(chan struct{})
	defer close(done)

	// set up - Applications and informers
	setup(t, svcConfigFile, done)

	// create objects in cluster
	deployObj := createObject(t, newUnstructuredDeployment(namespace, application, version))
	svcObj := createObject(t, newUnstructuredService(namespace, application, version))

	// creation of these objects should trigger handler which will add application to Applications map
	assert.Eventually(t, func() bool {
		return assertApplicationExists(t, namespace, application, []string{track}, []string{version})
	}, 10*time.Second, 100*time.Millisecond)

	assertV(t, version, deployObj)
	assertV(t, version, svcObj)

	// update version on first object (so mismatch)
	updateLabel(deployObj, versionLabel, version2)
	assertV(t, version2, deployObj)
	updateObject(t, deployObj)

	// eventually track map cleared
	assert.Eventually(t, func() bool { return assertApplicationExists(t, namespace, application, []string{}, []string{version}) }, 10*time.Second, 100*time.Millisecond)

	// update version on second object (so match)
	updateLabel(svcObj, versionLabel, version2)
	assertV(t, version2, svcObj)
	updateObject(t, svcObj)

	// eventually track reset
	assert.Eventually(t, func() bool {
		return assertApplicationExists(t, namespace, application, []string{track}, []string{version, version2})
	}, 10*time.Second, 100*time.Millisecond)
}

func assertV(t *testing.T, expected string, obj *unstructured.Unstructured) {
	assert.NotNil(t, obj)
	v, ok := getVersion(obj)
	assert.True(t, ok)
	assert.Equal(t, expected, v)
}

// test deletion of application object does not remove it from Applications; BUT the track map should be cleared
func TestDelete(t *testing.T) {
	var svcConfigFile = "../../testdata/abninputs/config.yaml"
	var namespace = "default"
	var application = "backend"
	var track = application
	var version = "v1"

	// set up - initialize channel
	done := make(chan struct{})
	defer close(done)

	// set up - Applications and informers
	setup(t, svcConfigFile, done)

	// create objects in cluster
	createObject(t, newUnstructuredDeployment(namespace, application, version))
	svcObj := createObject(t, newUnstructuredService(namespace, application, version))

	// creation of these objects should trigger handler which will add application to Applications map
	assert.Eventually(t, func() bool {
		return assertApplicationExists(t, namespace, application, []string{track}, []string{version})
	}, 10*time.Second, 100*time.Millisecond)

	// delete one of the objects
	deleteObject(t, svcObj)

	// eventually track map cleared
	assert.Eventually(t, func() bool { return assertApplicationExists(t, namespace, application, []string{}, []string{version}) }, 10*time.Second, 100*time.Millisecond)
}

func setup(t *testing.T, svcConfigFile string, done chan struct{}) {
	// set up - clear ApplicationsMap
	abnapp.Applications.Clear()

	// set up - define watcher and start informers
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	w := NewIter8Watcher(svcConfigFile)
	assert.NotNil(t, w)
	w.Start(done)
}

func assertApplicationExists(t *testing.T, namespace, application string, tracks []string, versions []string) bool {
	r := true

	a, err := abnapp.Applications.Get(namespace + "/" + application)
	r = r && assert.NoError(t, err)
	r = r && assert.NotNil(t, a)

	r = r && assertApplication(t, a, applicationAssertion{
		namespace: namespace,
		name:      application,
		tracks:    tracks,
		versions:  versions,
	})

	return r
}

func updateLabel(obj *unstructured.Unstructured, key string, value string) {
	if value == "" {
		delete((obj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{}), key)
	} else {
		(obj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[key] = value
	}
}

// for testing, way to get GVR from object
func getGVR(uObj *unstructured.Unstructured) *schema.GroupVersionResource {
	switch uObj.GetKind() {
	case "Service":
		return &schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "Deployment":
		return &schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	default:
		return nil
	}
}

func createObject(t *testing.T, uObj *unstructured.Unstructured) *unstructured.Unstructured {

	gvr := getGVR(uObj)
	assert.NotNil(t, gvr)

	createdObj, err := k8sclient.Client.Dynamic().
		Resource(*gvr).Namespace(uObj.GetNamespace()).
		Create(
			context.TODO(),
			uObj,
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)

	return createdObj
}

func updateObject(t *testing.T, uObj *unstructured.Unstructured) *unstructured.Unstructured {
	gvr := getGVR(uObj)
	assert.NotNil(t, gvr)

	updatedObj, err := k8sclient.Client.Dynamic().
		Resource(*gvr).Namespace(uObj.GetNamespace()).
		Update(
			context.TODO(),
			uObj,
			metav1.UpdateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, updatedObj)

	return updatedObj
}

func deleteObject(t *testing.T, uObj *unstructured.Unstructured) *unstructured.Unstructured {
	gvr := getGVR(uObj)
	assert.NotNil(t, gvr)

	err := k8sclient.Client.Dynamic().
		Resource(*gvr).Namespace(uObj.GetNamespace()).
		Delete(
			context.TODO(),
			uObj.GetName(),
			metav1.DeleteOptions{},
		)
	assert.NoError(t, err)

	return nil
}

func TestTrackNames(t *testing.T) {
	// also tests validObjectNames()
	for _, tt := range []struct {
		name          string
		numCandidates int
	}{
		{"zero", 0},
		{"one", 1},
		{"two", 2},
		{"ten", 10},
	} {
		t.Run(tt.name, func(t *testing.T) {
			names := getValidObjectNames(tt.name, tt.numCandidates)
			assert.Equal(t, tt.numCandidates+1, len(names))
			assert.Contains(t, names, tt.name)
			if tt.numCandidates > 0 {
				assert.Contains(t, names, fmt.Sprintf("%s-candidate-%d", tt.name, tt.numCandidates))
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	assert.True(t, containsString([]string{"one", "two", "three"}, "one"))
	assert.False(t, containsString([]string{"one", "two", "three"}, "four"))
}
