package application

import (
	"testing"
	"time"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/cli"
)

func TestApplicationNotInClusterRead(t *testing.T) {
	setup(t)
	a, err := Applications.readFromSecret("namespace/name")
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

func TestApplicationNotInClusterGet(t *testing.T) {
	setup(t)
	// not in cluster so created
	a, err := Applications.Get("namespace/name")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not found")

	assertApplication(t, a, applicationAssertion{
		namespace: "namespace",
		name:      "name",
		tracks:    []string{},
		versions:  []string{},
	})
}

func TestApplicationInCluster(t *testing.T) {
	setup(t)
	a, err := Applications.readFromSecret("default/application")
	assert.NoError(t, err)

	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	assertVersion(t, a.Versions["v1"], versionAssertion{
		track:   "",
		metrics: []string{"metric1"},
	})
	assertVersion(t, a.Versions["v2"], versionAssertion{
		track:   "candidate",
		metrics: []string{},
	})

	writeVerify(t, a)
}

func TestApplicationInClusterGet(t *testing.T) {
	setup(t)
	// in cluster, no error
	a, err := Applications.Get("default/application")
	assert.NoError(t, err)

	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})
}

func TestWrite(t *testing.T) {
	setup(t)

	a, _ := Applications.readFromSecret("default/application")
	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	// modify application in some way
	a.Tracks["foo"] = "v1"

	// Write writes immediately
	Applications.Write(a)
	b, _ := Applications.readFromSecret("default/application")
	// changed
	assertApplication(t, b, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate", "foo"},
		versions:  []string{"v1", "v2"},
	})
}

func TestWriteLimit(t *testing.T) {
	setup(t)
	BatchWriteInterval = time.Duration(0)
	maxApplicationDataBytes = 150

	a, err := Applications.readFromSecret("default/application")
	assert.NoError(t, err)
	assert.NotNil(t, a)

	// v1 is not associated with a track; only v2 is
	assert.NotEqual(t, len(a.Tracks), len(a.Versions))

	// because maxApplicationDatBytes is so small, should delete v1
	err = Applications.Write(a)
	assert.NoError(t, err)

	b, err := Applications.readFromSecret("default/application")
	assert.NoError(t, err)
	assert.NotNil(t, b)

	// only v2 is present
	assert.Equal(t, len(b.Tracks), len(b.Versions))
}

func TestBatchedWrite(t *testing.T) {
	setup(t)
	BatchWriteInterval = time.Duration(2 * time.Second)

	a, _ := Applications.Get("default/application")
	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	// modify application in some way
	a.Tracks["foo"] = "v1"

	// BatchedWrite should not write; too soon
	Applications.BatchedWrite(a)
	b, _ := Applications.readFromSecret("default/application")
	// no change; it has been too soon
	assertApplication(t, b, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	// time.Sleep(10 * time.Second)
	time.Sleep(BatchWriteInterval)

	// BatchedWrite should succeed; we waited > BatchWriteInterval
	Applications.BatchedWrite(a)
	c, _ := Applications.readFromSecret("default/application")
	// changed
	assertApplication(t, c, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate", "foo"},
		versions:  []string{"v1", "v2"},
	})
}

func TestFlush(t *testing.T) {
	setup(t)
	BatchWriteInterval = time.Duration(2 * time.Second)

	a, _ := Applications.Get("default/application")
	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	// modify application in some way
	a.Tracks["foo"] = "v1"

	// BatchedWrite should not write; too soon
	Applications.BatchedWrite(a)
	b, _ := Applications.readFromSecret("default/application")
	// no change; it has been too soon
	assertApplication(t, b, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	// avoid need to sleep by resetting BatchedWriteInterval
	BatchWriteInterval = time.Duration(0)

	// still not written since no second casll was made
	b, _ = Applications.readFromSecret("default/application")
	// no change; it has been too soon
	assertApplication(t, b, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	Applications.flush()

	// now will have been written
	c, _ := Applications.readFromSecret("default/application")
	// changed
	assertApplication(t, c, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate", "foo"},
		versions:  []string{"v1", "v2"},
	})
}

func TestPeriodicFlush(t *testing.T) {
	setup(t)
	BatchWriteInterval = time.Duration(1 * time.Second)
	flushMultiplier = 2

	done := make(chan struct{})
	Applications.PeriodicApplicationsFlush(done)

	a, _ := Applications.Get("default/application")
	assertApplication(t, a, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	// modify application in some way
	a.Tracks["foo"] = "v1"

	// don't write it now; wait for PeriodApplicationsFlush to trigger
	time.Sleep(time.Duration(1+flushMultiplier) * BatchWriteInterval)

	// should have been flushed
	c, _ := Applications.readFromSecret("default/application")
	// changed
	assertApplication(t, c, applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate", "foo"},
		versions:  []string{"v1", "v2"},
	})

	// terminate go flusher
	close(done)
}

func TestGetVersion(t *testing.T) {
	setup(t)
	a, _ := Applications.readFromSecret("default/application")

	var v *Version
	var isNew bool

	// get a version that exists
	v, _ = a.GetVersion("v1", true)

	assertVersion(t, v, versionAssertion{
		track:   "",
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
		Metrics: map[string]*SummaryMetric{},
	}
	assert.Nil(t, v.GetTrack())

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

func setup(t *testing.T) {
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	Applications.Clear()
	maxApplicationDataBytes = 750000
	yamlToSecret("../../testdata", "abninputs/readtest.yaml", "default/application")
}

func writeVerify(t *testing.T, a *Application) *Application {
	application := a.Name
	// write application to cluster (should create the secret, if not present)
	err := Applications.Write(a)
	assert.NoError(t, err)

	// verify can read it back
	a, err = Applications.readFromSecret(application)
	assert.NotNil(t, a)
	assert.NoError(t, err)
	return a
}
