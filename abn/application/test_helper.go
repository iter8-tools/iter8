package application

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/summarymetrics"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func yamlToSecret(folder, file, name string) error {
	err := ensureSecretCreated(name)
	if err != nil {
		return err
	}
	a, _ := yamlToApplication(name, folder, file)
	return Applications.Write(a)
}

func ensureSecretCreated(application string) error {
	namespace := namespaceFromKey(application)
	name := secretNameFromKey(application)
	_, err := k8sclient.Client.Typed().CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err = k8sclient.Client.Typed().CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	return err
}

func readYamlFromFile(folder, file string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	fname := filepath.Join(filepath.Dir(filename), folder, file)
	fn := filepath.Clean(fname)
	return os.ReadFile(fn)
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
		return NewApplication(name), nil
	}
	a.Name = name

	// Initialize versions if not already initialized
	if a.Versions == nil {
		a.Versions = Versions{}
	}
	for _, v := range a.Versions {
		if v.Metrics == nil {
			v.Metrics = map[string]*summarymetrics.SummaryMetric{}
		}
	}

	return a, nil
}

type applicationAssertion struct {
	namespace, name  string
	tracks, versions []string
}

func assertApplication(t *testing.T, a *Application, assertion applicationAssertion) bool {
	r := true
	r = r && assert.NotNil(t, a)
	r = r && assert.Contains(t, a.String(), assertion.namespace+"/"+assertion.name)

	namespace, name := splitApplicationKey(a.Name)
	r = r && assert.Equal(t, assertion.name, name)
	r = r && assert.Equal(t, assertion.namespace, namespace)

	r = r && assert.Len(t, a.Tracks, len(assertion.tracks))
	for _, track := range assertion.tracks {
		r = r && assert.Contains(t, a.Versions, a.Tracks[track])
	}
	r = r && assert.Len(t, a.Versions, len(assertion.versions))

	for _, v := range a.Versions {
		r = r && assert.NotNil(t, v.Metrics)
	}

	return r
}

type versionAssertion struct {
	track   string
	metrics []string
}

func assertVersion(t *testing.T, v *Version, assertion versionAssertion) bool {
	r := true

	r = r && assert.NotNil(t, v)
	r = r && assert.Contains(t, v.String(), "- metrics:")

	r = r && assert.Len(t, v.Metrics, len(assertion.metrics))
	r = r && assert.NotNil(t, v.Metrics)
	for m := range v.Metrics {
		r = r && assert.Contains(t, assertion.metrics, m)
	}
	return r
}

// Clear the application map
func (m *ThreadSafeApplicationMap) Clear() {
	m.mutex.Lock()
	m.apps = map[string]*Application{}
	m.lastWriteTimes = map[string]*time.Time{}
	m.mutexes = map[string]*sync.RWMutex{}
	m.mutex.Unlock()
}

// NumApplications asserts that the number of applications in the application map equals the given length
func NumApplications(t *testing.T, length int) bool {
	return assert.Len(t, Applications.apps, length)
}
