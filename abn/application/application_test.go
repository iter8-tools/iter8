package application

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplication(t *testing.T) {
	var a *Application
	var err error

	rw := ApplicationReaderWriter{Client: initKubeDriver(
		Manifest{
			folder:    "../../testdata",
			file:      "abninputs/readtest.yaml",
			name:      "application",
			namespace: "default",
		},
	).Clientset}

	// read application NOT in cluster
	a, err = rw.Read("namespace/name")
	assert.NotNil(t, a)
	assert.Equal(t, "name", a.Name)
	assert.Equal(t, "namespace", a.Namespace)
	assert.Len(t, a.Tracks, 0)
	assert.Len(t, a.Versions, 0)
	assert.ErrorContains(t, err, "not found")
	assert.Contains(t, a.String(), "Application namespace/name")

	// write application to cluster (should create the secret)
	err = a.Write()
	assert.NoError(t, err)
	// verify can read it back
	a, err = rw.Read("namespace/name")
	assert.NotNil(t, a)
	assert.NoError(t, err)

	// read another application IN cluster
	a, err = rw.Read("default/application")
	assert.NotNil(t, a)
	assert.NoError(t, err)
	assert.Equal(t, "application", a.Name)
	assert.Equal(t, "default", a.Namespace)
	assert.Len(t, a.Versions, 2)
	assert.Len(t, a.Tracks, 1)
	assert.Equal(t, "v2", a.Tracks["candidate"])

	var v *Version
	var isNew bool

	// get a version that exists
	v, isNew = a.GetVersion("v1", true)
	assert.NotNil(t, v)
	assert.False(t, isNew)
	assert.Len(t, v.History, 1)
	assert.False(t, v.IsReady())
	assert.Nil(t, v.GetTrack())
	assert.Contains(t, v.String(), "- history: [new]")

	// get a version that doesn't exist without allowing new creation
	v, isNew = a.GetVersion("foo", false)
	assert.Nil(t, v)
	assert.True(t, isNew)

	// get a version that doesn't exist allowing new creation
	v, isNew = a.GetVersion("foo", true)
	assert.NotNil(t, v)
	assert.True(t, isNew)
	assert.Len(t, v.History, 0)

	// write the application back to the cluster (should update the secret)
	err = a.Write()
	assert.NoError(t, err)
	a, err = rw.Read("default/application")
	assert.NotNil(t, a)
	assert.NoError(t, err)
}

type Manifest struct {
	folder    string
	file      string
	name      string
	namespace string
}

func initKubeDriver(secretManifests ...Manifest) *driver.KubeDriver {
	kd := driver.NewFakeKubeDriver(cli.New())
	for _, manifest := range secretManifests {
		byteArray, _ := ioutil.ReadFile(base.CompletePath(manifest.folder, manifest.file))
		s, _ := kd.Clientset.CoreV1().Secrets(manifest.namespace).Create(context.TODO(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      manifest.name,
				Namespace: manifest.namespace,
			},
			StringData: map[string]string{KEY: string(byteArray)},
		}, metav1.CreateOptions{})
		kd.Clientset.CoreV1().Secrets(manifest.namespace).Update(context.TODO(), s, metav1.UpdateOptions{})
	}
	return kd
}

func TestVersionAndSummaryMetric(t *testing.T) {
	var m *SummaryMetric
	var isNew bool

	v := &Version{
		History:             []VersionEvent{},
		Metrics:             map[string]*SummaryMetric{},
		LastUpdateTimestamp: time.Now(),
	}
	assert.Nil(t, v.GetTrack())
	assert.False(t, v.IsReady())

	v.AddEvent(VersionNewEvent)
	v.AddEvent(VersionReadyEvent)
	v.AddEvent(VersionMapTrackEvent, "track")
	assert.Equal(t, "track", *v.GetTrack())
	assert.True(t, v.IsReady())

	v.AddEvent(VersionNoLongerReadyEvent)
	v.AddEvent(VersionUnmapTrackEvent)
	assert.Nil(t, v.GetTrack())
	assert.False(t, v.IsReady())

	// test GetMetic w/o allowNew
	m, isNew = v.GetMetric("foo", false)
	assert.Nil(t, m)
	assert.False(t, isNew)

	// and with allowNew
	m, isNew = v.GetMetric("foo", true)
	assert.NotNil(t, m)
	assert.True(t, isNew)
	// new metric is empty
	assert.Equal(t, uint32(0), m.Count())
	assert.Equal(t, float64(0), m.Sum())

	// add values
	m.Add(float64(27))
	m.Add(float64(56))
	assert.Equal(t, uint32(2), m.Count())
	assert.Equal(t, float64(27), m.Min())
	assert.Equal(t, float64(56), m.Max())
	assert.Equal(t, float64(83), m.Sum())
	assert.Equal(t, float64(3865), m.SumSquares())
	assert.Equal(t, "[2] 27.000000, 56.000000, 83.000000, 3865.000000", m.String())

	// try again
	m, isNew = v.GetMetric("foo", false)
	assert.NotNil(t, m)
	assert.False(t, isNew)
	assert.Equal(t, uint32(2), m.Count())

	m, isNew = v.GetMetric("foo", true)
	assert.NotNil(t, m)
	assert.False(t, isNew)
	assert.Equal(t, uint32(2), m.Count())
	assert.Equal(t, float64(3865), m.SumSquares())
}
