package watcher

import (
	"testing"

	app "github.com/iter8-tools/iter8/abn/application"
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

	// setup: clear Applications
	Applications = map[string]*app.Application{}

	var wo WatchedObject

	// no name -- not added
	wo = newWatchedObject(nil, nil, nil, false, fakeKD)
	Add(wo)
	assert.Empty(t, Applications)

	// name but no version -- not added
	wo = newWatchedObject(&app1, nil, nil, false, fakeKD)
	Add(wo)
	assert.Empty(t, Applications)

	// name and version -- adds
	wo = newWatchedObject(&app1, &ver1, &trk1, false, fakeKD)
	Add(wo)
	// Update(wo, app1)
	// assert.Contains(t, Applications, app1)
	assert.Len(t, Applications, 1)
	a, _ := GetApplication(nn1, nil)
	assert.Len(t, a.Versions, 1)
	v, _ := a.GetVersion(ver1, false)
	assert.False(t, v.IsReady())
	// not ready, so track is ""
	assert.Nil(t, v.GetTrack())
	assert.Len(t, a.Tracks, 0)

	// add another same name, version, ready
	wo = newWatchedObject(&app1, &ver1, &trk2, true, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Applications, 1)
	a, _ = GetApplication(nn1, nil)
	assert.Len(t, a.Versions, 1)
	v, _ = a.GetVersion(ver1, false)
	assert.True(t, v.IsReady())
	// ready, so track is set
	assert.Equal(t, trk2, *v.GetTrack())
	assert.Len(t, a.Tracks, 1)

	// add another same name, different version
	wo = newWatchedObject(&app1, &ver2, nil, false, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Applications, 1)
	a, _ = GetApplication(nn1, nil)
	assert.Len(t, a.Versions, 2)
	v, _ = a.GetVersion(ver1, false)
	assert.True(t, v.IsReady())
	v, _ = a.GetVersion(ver2, false)
	assert.False(t, v.IsReady())

	// add another name
	wo = newWatchedObject(&app2, &ver3, nil, true, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Applications, 2)
	assert.Contains(t, Applications, nn1)
	assert.Contains(t, Applications, nn2)

	// add another name but watching it
	wo = newWatchedObject(&app2, &ver3, nil, true, fakeKD)
	Add(wo)
	Update(wo)
	assert.Len(t, Applications, 2)
	a1, _ := GetApplication(nn1, nil)
	assert.Len(t, a1.Versions, 2)
	a2, _ := GetApplication(nn2, nil)
	assert.Len(t, a2.Versions, 1) // there is a version of the other app
	v, _ = a1.GetVersion(ver1, false)
	assert.True(t, v.IsReady())
	v, _ = a1.GetVersion(ver2, false)
	assert.False(t, v.IsReady())
	v, _ = a2.GetVersion(ver3, false)
	assert.True(t, v.IsReady())
}

type AppProfile struct {
	numVersions int
	ready       bool
	numTracks   int
}

func TestDelete(t *testing.T) {
	notrecordedapp := "notrecorded"
	notrecordedversion := "notrecorded"

	// setup: clear Applications
	Applications = map[string]*app.Application{}

	// create initial set of objects
	Add(newWatchedObject(&app1, &ver1, &trk1, true, fakeKD))
	assert.Len(t, Applications, 1)
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
	assertApplications(t, 1, appsProfile)

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
	assertApplications(t, 1, appsProfile)

	// if deleted object has a ready annotation then
	//   set version not ready
	//   remove track
	Delete(newWatchedObject(&app1, &ver1, &trk1, true, fakeKD))
	appsProfile[nn1] = AppProfile{numVersions: 2, ready: false, numTracks: 1}
	assertApplications(t, 1, appsProfile)
}

func assertApplications(t *testing.T, numApplications int, expected map[string]AppProfile) {
	assert.Len(t, Applications, len(expected))
	for n, a := range expected {
		assert.Len(t, Applications[n].Versions, a.numVersions)
		assert.Len(t, Applications[n].Tracks, a.numTracks)
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
	wo := WatchedObject{
		Obj:    &unstructured.Unstructured{Object: obj},
		Writer: &app.ApplicationReaderWriter{Client: kd.Clientset},
	}

	return wo
}
