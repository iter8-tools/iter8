package application

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func yamlToSecret(folder, file, name string) error {
	a, _ := yamlToApplication(name, folder, file)
	return Applications.Write(a)
}

func readYamlFromFile(folder, file string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	fname := filepath.Join(filepath.Dir(filename), folder, file)

	return ioutil.ReadFile(fname)
}

func yamlToApplication(name, folder, file string) (*Application, error) {
	byteArray, err := readYamlFromFile(folder, file)
	if err != nil {
		return nil, err
	}

	return byteArrayToApplication(name, byteArray)
}

func byteArrayToApplication(name string, data []byte) (*Application, error) {
	a := &Application{}
	err := yaml.Unmarshal(data, a)
	if err != nil {
		return &Application{
			Name:     name,
			Versions: Versions{},
			Tracks:   Tracks{},
		}, nil
	}
	a.Name = name

	// Initialize versions if not already initialized
	if a.Versions == nil {
		a.Versions = Versions{}
	}
	for _, v := range a.Versions {
		if v.Metrics == nil {
			v.Metrics = map[string]*SummaryMetric{}
		}
	}

	return a, nil
}

type applicationAssertion struct {
	namespace, name  string
	tracks, versions []string
}

func assertApplication(t *testing.T, a *Application, assertion applicationAssertion) {
	assert.NotNil(t, a)
	assert.Contains(t, a.String(), assertion.namespace+"/"+assertion.name)

	namespace, name := splitApplicationKey(a.Name)
	assert.Equal(t, assertion.name, name)
	assert.Equal(t, assertion.namespace, namespace)

	assert.Len(t, a.Tracks, len(assertion.tracks))
	for _, track := range assertion.tracks {
		assert.Contains(t, a.Versions, a.Tracks[track])
	}
	assert.Len(t, a.Versions, len(assertion.versions))

	for _, v := range a.Versions {
		assert.NotNil(t, v.Metrics)
	}
}

type versionAssertion struct {
	track   string
	metrics []string
}

func assertVersion(t *testing.T, v *Version, assertion versionAssertion) {
	assert.NotNil(t, v)

	assert.Len(t, v.Metrics, len(assertion.metrics))
	assert.NotNil(t, v.Metrics)
	for m := range v.Metrics {
		assert.Contains(t, assertion.metrics, m)
	}
}

// Clear the application map
func (m *ThreadSafeApplicationMap) Clear() {
	m.mutex.Lock()
	m.apps = map[string]*Application{}
	m.lastWriteTimes = map[string]*time.Time{}
	m.mutexes = map[string]*sync.RWMutex{}
	m.mutex.Unlock()
}

func NumApplications(t *testing.T, length int) {
	assert.Len(t, Applications.apps, length)
}
