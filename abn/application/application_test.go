package application

import (
	"testing"
	"time"

	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestApplicationNotInCluster(t *testing.T) {
	rw := setup(t)
	a, err := rw.Read("namespace/name")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not found")

	assertApplication(t, a, applicationAssertion{
		namespace: "namespace",
		name:      "name",
		tracks:    []string{},
		versions:  []string{},
	})

	writeVerify(t, a)
}

func TestApplicationInCluster(t *testing.T) {
	rw := setup(t)
	a, err := rw.Read("default/application")
	assert.NoError(t, err)

	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	assertVersion(t, a.Versions["v1"], versionAssertion{
		events:  []VersionEventType{VersionNewEvent},
		track:   "",
		ready:   false,
		metrics: []string{"metric1"},
	})
	assertVersion(t, a.Versions["v2"], versionAssertion{
		events:  []VersionEventType{VersionNewEvent, VersionReadyEvent, VersionMapTrackEvent},
		track:   "candidate",
		ready:   true,
		metrics: []string{},
	})

	writeVerify(t, a)
}

func TestGetVersion(t *testing.T) {
	rw := setup(t)
	a, _ := rw.Read("default/application")

	var v *Version
	var isNew bool

	// get a version that exists
	v, _ = a.GetVersion("v1", true)

	assertVersion(t, v, versionAssertion{
		events:  []VersionEventType{VersionNewEvent},
		track:   "",
		ready:   false,
		metrics: []string{"metric1"},
	})

	// get a version that doesn't exist without allowing new creation
	v, isNew = a.GetVersion("foo", false)
	assert.Nil(t, v)
	assert.True(t, isNew)

	// get a version that doesn't exist allowing new creation
	v, isNew = a.GetVersion("foo", true)
	assert.NotNil(t, v)
	assert.True(t, isNew)

	a = writeVerify(t, a)
	// verify version foo is now present
	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2", "foo"},
	})
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

func setup(t *testing.T) *ApplicationReaderWriter {
	kd := driver.NewFakeKubeDriver(cli.New())
	YamlToSecret("../../testdata", "abninputs/readtest.yaml", "default", "application", kd)
	return &ApplicationReaderWriter{Client: kd.Clientset}
}

func writeVerify(t *testing.T, a *Application) *Application {
	application := a.Namespace + "/" + a.Name
	// write application to cluster (should create the secret, if not present)
	err := a.Write()
	assert.NoError(t, err)

	// verify can read it back
	a, err = a.ReaderWriter.Read(application)
	assert.NotNil(t, a)
	assert.NoError(t, err)
	return a
}
