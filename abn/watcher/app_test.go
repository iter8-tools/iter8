package watcher

import (
	"testing"

	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	objName      = "myobj"
	objNamespace = "mynamespace"
)

var app1 string = "myapp-one"
var app2 string = "myapp-two"
var nn1 string = objNamespace + "/" + app1
var nn2 string = objNamespace + "/" + app2
var ver1 string = "version1"
var ver2 string = "version2"
var ver3 string = "version3"
var trk1 string = "track1"
var trk2 string = "track2"
var fakeKD *driver.KubeDriver = driver.NewFakeKubeDriver(cli.New())

func TestAddUpdate(t *testing.T) {
	// // check track

	// setup: clear Apps
	Apps = map[string]Application{}

	var wo WatchedObject

	// no name -- not added
	wo = newWatchedObject(nil, nil, nil, false, fakeKD)
	Add(wo)
	assert.Empty(t, Apps)

	// name but no version -- not added
	wo = newWatchedObject(&app1, nil, nil, false, fakeKD)
	Add(wo)
	assert.Empty(t, Apps)

	// name and version -- adds
	wo = newWatchedObject(&app1, &ver1, &trk1, false, fakeKD)
	Add(wo)
	// Update(wo, app1)
	// assert.Contains(t, Apps, app1)
	assert.Len(t, Apps, 1)
	assert.Len(t, Apps[nn1].Versions, 1)
	assert.False(t, Apps[nn1].Versions[ver1].Ready)
	// not ready, so track is ""
	assert.Equal(t, "", Apps[nn1].Versions[ver1].Track)
	assert.Len(t, Apps[nn1].Tracks, 0)

	// add another same name, version, ready
	wo = newWatchedObject(&app1, &ver1, &trk2, true, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Apps, 1)
	assert.Len(t, Apps[nn1].Versions, 1)
	assert.True(t, Apps[nn1].Versions[ver1].Ready)
	// ready, so track is set
	assert.Equal(t, trk2, Apps[nn1].Versions[ver1].Track)
	assert.Len(t, Apps[nn1].Tracks, 1)

	// add another same name, different version
	wo = newWatchedObject(&app1, &ver2, nil, false, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Apps, 1)
	assert.Len(t, Apps[nn1].Versions, 2)
	assert.True(t, Apps[nn1].Versions[ver1].Ready)
	assert.False(t, Apps[nn1].Versions[ver2].Ready)

	// add another name
	wo = newWatchedObject(&app2, &ver3, nil, true, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Apps, 2)
	assert.Contains(t, Apps, nn1)
	assert.Contains(t, Apps, nn2)

	// add another name but watching it
	wo = newWatchedObject(&app2, &ver3, nil, true, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Apps, 2)
	assert.Len(t, Apps[nn1].Versions, 2)
	assert.Len(t, Apps[nn2].Versions, 1) // there is a version of the other app
	assert.True(t, Apps[nn1].Versions[ver1].Ready)
	assert.False(t, Apps[nn1].Versions[ver2].Ready)
	assert.True(t, Apps[nn2].Versions[ver3].Ready)
}

type AppProfile struct {
	numVersions int
	ready       bool
	numTracks   int
}

func TestDelete(t *testing.T) {
	notrecordedapp := "notrecorded"
	notrecordedversion := "notrecorded"

	// setup: clear Apps
	Apps = map[string]Application{}

	// create initial set of objects
	Add(newWatchedObject(&app1, &ver1, &trk1, true, fakeKD))
	assert.Len(t, Apps, 1)
	Add(newWatchedObject(&app1, &ver2, &trk2, true, fakeKD))
	Add(newWatchedObject(&app2, &ver1, &trk1, true, fakeKD))
	// base assertions
	appsProfile := map[string]AppProfile{
		nn1: {
			numVersions: 2,
			ready:       true,
			numTracks:   2,
		},
		nn2: {
			numVersions: 1,
			ready:       true,
			numTracks:   1,
		},
	}
	assertApps(t, 1, appsProfile)

	// no name --> ignore
	Delete(newWatchedObject(nil, nil, nil, false, fakeKD))
	// name but no version --> ignore
	Delete(newWatchedObject(&app1, nil, nil, false, fakeKD))
	// has name but no information recorded
	// should not happen but we are deleting so just ignore
	Delete(newWatchedObject(&notrecordedapp, &ver1, nil, false, fakeKD))
	// has known name and unrecognized version
	// should not happend but we are deleting so just ignore
	Delete(newWatchedObject(&app1, &notrecordedversion, nil, false, fakeKD))
	assertApps(t, 1, appsProfile)

	// if deleted object has a ready annotation then
	//   set version not ready
	//   remove track
	Delete(newWatchedObject(&app1, &ver1, &trk1, true, fakeKD))
	appsProfile[nn1] = AppProfile{numVersions: 2, ready: false, numTracks: 1}
	assertApps(t, 1, appsProfile)
}

func assertApps(t *testing.T, numApps int, expected map[string]AppProfile) {
	assert.Len(t, Apps, len(expected))
	for n, a := range expected {
		assert.Len(t, Apps[n].Versions, a.numVersions)
		assert.Len(t, Apps[n].Tracks, a.numTracks)
	}
}

func newWatchedObject(name *string, version *string, track *string, ready bool, kd *driver.KubeDriver) WatchedObject {
	labels := map[string]string{}
	if name != nil {
		labels[NAME_LABEL] = *name
	}
	if version != nil {
		labels[VERSION_LABEL] = *version
	}
	annotations := map[string]string{}
	if track != nil {
		annotations[TRACK_ANNOTATION] = *track
	}
	if ready {
		annotations[READY_ANNOTATION] = "true"
	}

	o := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        objName,
			Namespace:   objNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{},
	}
	obj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&o)
	wo := WatchedObject{Obj: &unstructured.Unstructured{Object: obj}, Driver: kd}

	return wo
}
