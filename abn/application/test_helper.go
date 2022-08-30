package application

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func YamlToApplication(name, folder, file string) (*Application, error) {
	byteArray, err := readYamlFromFile(folder, file)
	if err != nil {
		return nil, err
	}

	return byteArrayToApplication(name, byteArray)
}

func YamlToSecret(folder, file, name string, kd *driver.KubeDriver) error {
	byteArray, err := readYamlFromFile(folder, file)
	if err != nil {
		return err
	}

	secretName := GetSecretNameFromKey(name)
	secretNamespace := GetNamespaceFromKey(name)

	_, err = kd.Clientset.CoreV1().Secrets(secretNamespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		StringData: map[string]string{KEY: string(byteArray)},
	}, metav1.CreateOptions{})
	return err
}

func readYamlFromFile(folder, file string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	fname := filepath.Join(filepath.Dir(filename), folder, file)

	return ioutil.ReadFile(fname)
}

func byteArrayToApplication(name string, data []byte) (*Application, error) {
	a := &Application{}
	err := yaml.Unmarshal(data, a)
	if err != nil {
		return NewApplication(name), nil
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

	assert.Equal(t, assertion.name, GetNameFromKey(a.Name))
	assert.Equal(t, assertion.namespace, GetNamespaceFromKey(a.Name))

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
	ready   bool
	metrics []string
}

func assertVersion(t *testing.T, v *Version, assertion versionAssertion) {
	assert.NotNil(t, v)
	assert.Contains(t, v.String(), strconv.FormatBool(assertion.ready))

	track := v.GetTrack()
	if assertion.track == "" {
		assert.Nil(t, track)
	} else {
		assert.Equal(t, assertion.track, *track)
	}

	assert.Equal(t, assertion.ready, v.IsReady())

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

func NoApplications(t *testing.T) {
	assert.Empty(t, Applications.apps)
}

func Len(t *testing.T, length int) {
	assert.Len(t, Applications.apps, length)
}

func Contains(t *testing.T, application string) {
	assert.Contains(t, Applications.apps, application)
}

func NotContains(t *testing.T, application string) {
	assert.NotContains(t, Applications.apps, application)
}
