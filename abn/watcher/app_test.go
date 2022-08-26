package watcher

import (
	"testing"

	abnapp "github.com/iter8-tools/iter8/abn/application"
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
var T string = "true"
var F string = "false"

func setup() {
	abnapp.Applications.Clear()
	abnapp.Applications.SetReaderWriter(&abnapp.ApplicationReaderWriter{Client: driver.NewFakeKubeDriver(cli.New()).Clientset})
}

func TestAddUpdate(t *testing.T) {
	setup()
	abnapp.NoApplications(t)

	var wo WatchedObject

	// no name -- not added
	wo = newWatchedObject(nil, nil, nil, nil)
	Add(wo)
	abnapp.NoApplications(t)

	// name but no version -- not added
	wo = newWatchedObject(&app1, nil, nil, nil)
	Add(wo)
	abnapp.NoApplications(t)

	// name and version -- adds
	wo = newWatchedObject(&app1, &ver1, &trk1, nil)
	Add(wo)
	// Update(wo, app1)
	// assert.Contains(t, Applications, app1)
	abnapp.Len(t, 1)

	a, _ := abnapp.Applications.Get(nn1, true)
	assert.Len(t, a.Versions, 1)
	v, _ := a.GetVersion(ver1, false)
	assert.False(t, v.IsReady())
	// not ready, so track is ""
	assert.Nil(t, v.GetTrack())
	assert.Len(t, a.Tracks, 0)

	// add another same name, version, ready
	wo = newWatchedObject(&app1, &ver1, &trk2, &T)
	Add(wo)
	Update(wo)
	abnapp.Len(t, 1)
	a, _ = abnapp.Applications.Get(nn1, true)
	assert.Len(t, a.Versions, 1)
	v, _ = a.GetVersion(ver1, false)
	assert.True(t, v.IsReady())
	// ready, so track is set
	assert.Equal(t, trk2, *v.GetTrack())
	assert.Len(t, a.Tracks, 1)

	// add another same name, version but not explicitly NOT ready
	// expect version to no longer be ready and not tracked
	wo = newWatchedObject(&app1, &ver1, nil, &F)
	Add(wo)
	a, _ = abnapp.Applications.Get(nn1, true)
	assert.Len(t, a.Versions, 1)
	v, _ = a.GetVersion(ver1, false)
	assert.False(t, v.IsReady())
	assert.Len(t, a.Tracks, 0)

	// add another same name, different version
	wo = newWatchedObject(&app1, &ver2, nil, nil)
	Add(wo)
	Update(wo)
	abnapp.Len(t, 1)
	a, _ = abnapp.Applications.Get(nn1, true)
	assert.Len(t, a.Versions, 2)
	v, _ = a.GetVersion(ver1, false)
	assert.False(t, v.IsReady()) // remains false
	v, _ = a.GetVersion(ver2, false)
	assert.False(t, v.IsReady())

	// add another name
	wo = newWatchedObject(&app2, &ver3, nil, &T)
	Add(wo)
	Update(wo)
	abnapp.Len(t, 2)
	abnapp.Contains(t, nn1)
	abnapp.Contains(t, nn2)

	// add another name but watching it
	wo = newWatchedObject(&app2, &ver3, nil, &T)
	Add(wo)
	Update(wo)
	abnapp.Len(t, 2)
	a1, _ := abnapp.Applications.Get(nn1, true)
	assert.Len(t, a1.Versions, 2)
	a2, _ := abnapp.Applications.Get(nn2, true)
	assert.Len(t, a2.Versions, 1) // there is a version of the other app
	v, _ = a1.GetVersion(ver1, false)
	assert.False(t, v.IsReady())
	v, _ = a1.GetVersion(ver2, false)
	assert.False(t, v.IsReady())
	v, _ = a2.GetVersion(ver3, false)
	assert.True(t, v.IsReady())
}

func TestDelete(t *testing.T) {
	notrecordedapp := "notrecorded"
	notrecordedversion := "notrecorded"

	setup()

	abnapp.Len(t, 0)

	// create initial set of objects
	Add(newWatchedObject(&app1, &ver1, &trk1, &T))
	abnapp.Len(t, 1)
	abnapp.Contains(t, nn1)
	abnapp.NotContains(t, nn2)

	Add(newWatchedObject(&app1, &ver2, &trk2, &T))
	abnapp.Len(t, 1)

	Add(newWatchedObject(&app2, &ver1, &trk1, &T))
	abnapp.Len(t, 2)
	abnapp.Contains(t, nn2)

	a1, err := abnapp.Applications.Get(nn1, true)
	assert.NoError(t, err)
	assert.NotNil(t, a1)
	assertApplication(t, a1, 2, 2)
	a2, err := abnapp.Applications.Get(nn2, true)
	assert.NoError(t, err)
	assertApplication(t, a2, 1, 1)
	//assertApplications(t, 2, appsProfile)

	// no name --> ignore
	Delete(newWatchedObject(nil, nil, nil, nil))
	// name but no version --> ignore
	Delete(newWatchedObject(&app1, nil, nil, nil))
	// has name but no information recorded
	// should not happen but we are deleting so just ignore
	Delete(newWatchedObject(&notrecordedapp, &ver1, nil, nil))
	// has known name and unrecognized version
	// should not happend but we are deleting so just ignore
	Delete(newWatchedObject(&app1, &notrecordedversion, nil, nil))

	// validate no changes
	a1, err = abnapp.Applications.Get(nn1, true)
	assert.NoError(t, err)
	assertApplication(t, a1, 2, 2)
	a2, err = abnapp.Applications.Get(nn2, true)
	assert.NoError(t, err)
	assertApplication(t, a2, 1, 1)

	// if deleted object has a ready annotation then
	//   set version not ready
	//   remove track
	Delete(newWatchedObject(&app1, &ver1, &trk1, &T))
	a1, err = abnapp.Applications.Get(nn1, true)
	assert.NoError(t, err)
	assertApplication(t, a1, 2, 1)
}

// Utility methods

func assertApplication(t *testing.T, a *abnapp.Application, nVersions int, nTracks int) {
	assert.NotNil(t, a)
	assert.Len(t, (*a).Versions, nVersions)
	assert.Len(t, a.Tracks, nTracks)
}

func newWatchedObject(name *string, version *string, track *string, ready *string) WatchedObject {
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
	if ready != nil {
		annotations[READY_ANNOTATION] = *ready
	}
	annotations[ITER8_ANNOTATION] = "true"

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
	wo := WatchedObject{Obj: &unstructured.Unstructured{Object: obj}}

	return wo
}
