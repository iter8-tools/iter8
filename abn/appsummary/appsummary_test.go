package appsummary

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	app         string = "default"
	new_app     string = "default/new_app"
	version     string = "v1"
	new_version string = "new_version"
	metric      string = "metric1"
	new_metric         = "new_metric"
)

func setup() *driver.KubeDriver {
	kd := driver.NewFakeKubeDriver(cli.New())
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../../testdata", "abninputs/readtest.yaml"))
	s, _ := kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app,
			Namespace: "default",
		},
		StringData: map[string]string{"versionData.yaml": string(byteArray)},
	}, metav1.CreateOptions{})
	s.ObjectMeta.Labels = map[string]string{"foo": "bar"}
	kd.Clientset.CoreV1().Secrets("default").Update(context.TODO(), s, metav1.UpdateOptions{})
	return kd
}

func (driver *MetricDriver) getMetric(application string, version string, metric string) (SummaryMetric, error) {
	a, err := driver.ReadApplicationSummary(application)
	if err != nil {
		return EmptySummaryMetric(), err
	}
	v, err := a.GetVersion(version)
	if err != nil {
		return EmptySummaryMetric(), err
	}
	m, err := v.GetMetric(metric)
	if err != nil {
		return EmptySummaryMetric(), err
	}

	return m, nil
}

func (driver *MetricDriver) getHistory(application string, version string) ([]VersionEvent, error) {
	a, err := driver.ReadApplicationSummary(application)
	if err != nil {
		return []VersionEvent{}, err
	}
	v, err := a.GetVersion(version)
	if err != nil {
		return []VersionEvent{}, err
	}

	return v.Data.History, nil
}

func TestAddMetric(t *testing.T) {
	kd := setup()

	var s SummaryMetric
	var err error
	var value float64 = 30.0

	md := MetricDriver{Client: kd.Clientset}

	// new app; it, version, and metric get created
	err = md.AddMetric(new_app, version, metric, value)
	assert.NoError(t, err)

	s, err = md.getMetric(app, version, metric)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), s.Count())

	// new version; it and metric get created
	err = md.AddMetric(app, new_version, metric, value)
	assert.NoError(t, err)
	s, err = md.getMetric(app, new_version, metric)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), s.Count())
	assert.Equal(t, value, s.Min())

	// new metric; is created
	err = md.AddMetric(app, version, new_metric, value)
	assert.NoError(t, err)
	s, err = md.getMetric(app, "v1", new_metric)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), s.Count())
	assert.Equal(t, value, s.Min())

	// existing metric; now see increase in count
	err = md.AddMetric(app, version, metric, value)
	assert.NoError(t, err)
	s, err = md.getMetric(app, version, metric)
	assert.NoError(t, err)
	assert.Equal(t, uint32(2), s.Count())
	assert.Equal(t, float64(75), s.Sum())
	assert.Equal(t, float64(30), s.Min())
}

func TestRecordEvent(t *testing.T) {
	kd := setup()

	var err error
	var history []VersionEvent

	md := MetricDriver{Client: kd.Clientset}

	// new app; it and version are created
	err = md.RecordEvent(new_app, version, VersionNewEvent)
	assert.NoError(t, err)
	history, err = md.getHistory(new_app, version)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(history))

	// new version; is created
	err = md.RecordEvent(app, new_version, VersionNewEvent)
	assert.NoError(t, err)
	history, _ = md.getHistory(app, new_version)
	assert.Equal(t, 1, len(history))

	// new metric; existing version updated
	err = md.RecordEvent(app, version, VersionMapTrackEvent, "track")
	assert.NoError(t, err)
	history, _ = md.getHistory(app, version)
	assert.Equal(t, 4, len(history))
}

func TestReadApplication(t *testing.T) {
}

func TestWriteApplication(t *testing.T) {
}
