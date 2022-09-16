package application

import (
	"testing"
	"time"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/cli"
)

func TestGetAndRead(t *testing.T) {
	app := applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	}
	defaultApp := applicationAssertion{
		namespace: "default",
		name:      "application",
		tracks:    []string{},
		versions:  []string{},
	}
	scenarios := map[string]struct {
		fetch         func(name string) (*Application, error)
		setup         func(t *testing.T, applications ...applicationSource)
		errorContains string
		isNil         bool
		appSpec       *applicationAssertion
	}{
		"GET in memory, in cluster":          {setup: setupInMemoryInCluster, fetch: Applications.Get, isNil: false, appSpec: &app, errorContains: ""},
		"READ in memory, in cluster":         {setup: setupInMemoryInCluster, fetch: Applications.Read, isNil: false, appSpec: &app, errorContains: ""},
		"GET in memory, not in cluster":      {setup: setupInMemoryNotInCluster, fetch: Applications.Get, isNil: false, appSpec: &app, errorContains: ""},
		"READ in memory, not in cluster":     {setup: setupInMemoryNotInCluster, fetch: Applications.Read, isNil: false, appSpec: &app, errorContains: ""},
		"GET not in memory, in cluster":      {setup: setupNotInMemoryInCluster, fetch: Applications.Get, isNil: true, appSpec: nil, errorContains: "not in memory"},
		"READ not in memory, in cluster":     {setup: setupNotInMemoryInCluster, fetch: Applications.Read, isNil: false, appSpec: &app, errorContains: ""},
		"GET not in memory, not in cluster":  {setup: setupNotInMemoryNotInCluster, fetch: Applications.Get, isNil: true, appSpec: nil, errorContains: "not in memory"},
		"READ not in memory, not in cluster": {setup: setupNotInMemoryNotInCluster, fetch: Applications.Read, isNil: false, appSpec: &defaultApp, errorContains: "not found"},
	}

	for label, s := range scenarios {
		t.Run(label, func(t *testing.T) {
			s.setup(t, applicationSource{namespace: "default", name: "application", folder: testdata, file: testfile})
			a, err := s.fetch("default/application")
			if s.errorContains != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, s.errorContains)
			} else {
				assert.NoError(t, err)
			}
			if s.appSpec == nil {
				assert.Nil(t, a)
			} else {
				assert.NotNil(t, a)
				assertApplication(t, a, *s.appSpec)
				writeVerify(t, a)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	setupInMemoryInCluster(t, applicationSource{namespace: "default", name: "application", folder: testdata, file: testfile})

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
	_ = Applications.Write(a)
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
	setupInMemoryInCluster(t, applicationSource{namespace: "default", name: "application", folder: testdata, file: testfile})
	BatchWriteInterval = time.Duration(0)
	maxApplicationDataBytes = 100

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
	ns := "test"
	nm := "batchedwrite"
	setupInMemoryInCluster(t, applicationSource{namespace: ns, name: nm, folder: testdata, file: testfile})
	BatchWriteInterval = time.Duration(2 * time.Second)

	a, _ := Applications.Get(ns + "/" + nm)
	assertApplication(t, a, applicationAssertion{
		namespace: ns,
		name:      nm,
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	// modify application in some way
	a.Tracks["foo"] = "v1"

	// BatchedWrite should not write; too soon
	_ = Applications.BatchedWrite(a)
	b, _ := Applications.readFromSecret(ns + "/" + nm)
	// no change; it has been too soon
	assertApplication(t, b, applicationAssertion{
		namespace: ns,
		name:      nm,
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	time.Sleep((250 * time.Millisecond) + BatchWriteInterval)

	// BatchedWrite should succeed; we waited > BatchWriteInterval
	_ = Applications.BatchedWrite(a)
	c, _ := Applications.Read(ns + "/" + nm)
	// changed
	assertApplication(t, c, applicationAssertion{
		namespace: ns,
		name:      nm,
		tracks:    []string{"candidate", "foo"},
		versions:  []string{"v1", "v2"},
	})
}

const testdata string = "../../testdata"
const testfile string = "abninputs/readtest.yaml"

func TestGetVersion(t *testing.T) {
	ns := "test"
	nm := "getversion"
	setupInMemoryInCluster(t, applicationSource{namespace: ns, name: nm, folder: testdata, file: testfile})
	a, _ := Applications.Read(ns + "/" + nm)
	// verify it is as expected
	assertApplication(t, a, applicationAssertion{
		namespace: ns,
		name:      nm,
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2"},
	})

	var v *Version
	var isNew bool

	// get a version that exists
	v, _ = a.GetVersion("v1", false)
	assertVersion(t, v, versionAssertion{
		track:   "",
		metrics: []string{"metric1"},
	})

	// get a version that doesn't exist without allowing new creation
	v, isNew = a.GetVersion("foo", false)
	assert.Nil(t, v)
	assert.False(t, isNew)

	// get a version that doesn't exist allowing new creation
	v, isNew = a.GetVersion("foo", true)
	assert.NotNil(t, v)
	assert.True(t, isNew)
	assertApplication(t, a, applicationAssertion{
		namespace: ns,
		name:      nm,
		tracks:    []string{"candidate"},
		versions:  []string{"v1", "v2", "foo"},
	})

	// b := writeVerify(t, a)

	application := a.Name
	// write application to cluster (should create the secret, if not present)
	err := Applications.Write(a)
	assert.NoError(t, err)
	// read back from cluster; clear Applications first so actually read from secret
	Applications.Clear()
	a, err = Applications.Read(application)
	assert.NotNil(t, a)
	assert.NoError(t, err)

	// verify version foo is now present
	assertApplication(t, a, applicationAssertion{
		namespace: ns,
		name:      nm,
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

type applicationSource struct {
	namespace string
	name      string
	folder    string
	file      string
}

func setupInitialization(t *testing.T) {
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	Applications.Clear()
	maxApplicationDataBytes = defaultMaxApplicationDataBytes
	BatchWriteInterval = defaultBatchWriteInterval
}

func setupNotInMemoryNotInCluster(t *testing.T, applications ...applicationSource) {
	setupInitialization(t)
}

func setupNotInMemoryInCluster(t *testing.T, applications ...applicationSource) {
	setupInitialization(t)
	for _, aSrc := range applications {
		_ = yamlToSecret(aSrc.folder, aSrc.file, aSrc.namespace+"/"+aSrc.name)
	}
}

func setupInMemoryInCluster(t *testing.T, applications ...applicationSource) {
	setupNotInMemoryInCluster(t, applications...)
	for _, aSrc := range applications {
		a, err := Applications.Read(aSrc.namespace + "/" + aSrc.name)
		assert.NoError(t, err)
		assert.NotNil(t, a)
	}
}

func setupInMemoryNotInCluster(t *testing.T, applications ...applicationSource) {
	setupInMemoryInCluster(t, applications...)
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
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
