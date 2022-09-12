package application

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func yamlToSecret(folder, file, name string) error {
	byteArray, err := readYamlFromFile(folder, file)
	if err != nil {
		return err
	}

	secretName := secretNameFromKey(name)
	secretNamespace := namespaceFromKey(name)

	_, err = k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		StringData: map[string]string{secretKey: string(byteArray)},
	}, metav1.CreateOptions{})
	return err
}

func readYamlFromFile(folder, file string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	fname := filepath.Join(filepath.Dir(filename), folder, file)

	return ioutil.ReadFile(fname)
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
