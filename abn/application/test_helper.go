package application

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	k8sdriver "github.com/iter8-tools/iter8/base/k8sdriver"
	metrics "github.com/iter8-tools/iter8/base/metrics"
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

func YamlToSecret(folder, file, namespace, name string, kd *k8sdriver.KubeDriver) error {
	byteArray, err := readYamlFromFile(folder, file)
	if err != nil {
		return err
	}

	_, err = kd.Clientset.CoreV1().Secrets(namespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
	a := GetNewApplication(name, nil)

	var versions Versions
	err := yaml.Unmarshal(data, &versions)
	if err != nil {
		return a, nil
	}

	// set Versions
	a.Versions = versions

	// initialize Tracks
	for version, v := range versions {
		track := v.GetTrack()
		if track != nil {
			a.Tracks[*track] = version
		}
		if v.History == nil {
			v.History = []VersionEvent{}
		}
		if v.Metrics == nil {
			v.Metrics = map[string]*metrics.SummaryMetric{}
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

	assert.Equal(t, assertion.name, a.Name)
	assert.Equal(t, assertion.namespace, a.Namespace)

	assert.Len(t, a.Tracks, len(assertion.tracks))
	for _, track := range assertion.tracks {
		assert.Contains(t, a.Versions, a.Tracks[track])
	}
	assert.Len(t, a.Versions, len(assertion.versions))

	for _, v := range a.Versions {
		assert.NotNil(t, v.History)
		assert.NotNil(t, v.Metrics)
	}
}

type versionAssertion struct {
	events  []VersionEventType
	track   string
	ready   bool
	metrics []string
}

func assertVersion(t *testing.T, v *Version, assertion versionAssertion) {
	assert.NotNil(t, v)
	assert.Contains(t, v.String(), "- history:")

	track := v.GetTrack()
	if assertion.track == "" {
		assert.Nil(t, track)
	} else {
		assert.Equal(t, assertion.track, *track)
	}

	assert.Equal(t, assertion.ready, v.IsReady())

	assert.Len(t, v.History, len(assertion.events))
	assert.NotNil(t, v.History)
	for i, e := range v.History {
		assert.Equal(t, assertion.events[i], e.Type)
	}

	assert.Len(t, v.Metrics, len(assertion.metrics))
	assert.NotNil(t, v.Metrics)
	for m := range v.Metrics {
		assert.Contains(t, assertion.metrics, m)
	}
}
