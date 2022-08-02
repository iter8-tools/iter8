package metricstore

import (
	"context"
	"io/ioutil"
	"math"
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
	new_app     string = "new_app"
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

func teardown() {

}

func TestGetSummaryMetric(t *testing.T) {
	kd := setup()
	defer teardown()

	var ms *MetricStoreSecret
	var s SummaryMetric
	var err error

	ms, err = NewMetricStoreSecret(new_app, kd)
	assert.NoError(t, err)

	// new app
	s, err = ms.GetSummaryMetric(metric, version)
	assert.ErrorContains(t, err, "no secret for application")
	assert.Equal(t, uint32(0), s.Count())
	assert.Equal(t, math.MaxFloat64, s.Min())

	// get new ms using valid app name
	ms, err = NewMetricStoreSecret(app, kd)
	assert.NoError(t, err)

	// new version
	s, err = ms.GetSummaryMetric(metric, new_version)
	assert.ErrorContains(t, err, "no data found for version")
	assert.Equal(t, uint32(0), s.Count())
	assert.Equal(t, math.MaxFloat64, s.Min())

	// new metric
	s, err = ms.GetSummaryMetric(new_metric, version)
	assert.ErrorContains(t, err, "no value found")
	assert.Equal(t, uint32(0), s.Count())
	assert.Equal(t, math.MaxFloat64, s.Min())

	// value available
	s, err = ms.GetSummaryMetric(metric, version)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), s.Count())
	assert.Equal(t, float64(45), s.Sum())
	assert.Equal(t, float64(45), s.Min())
}

func TestAddMetric(t *testing.T) {
	kd := setup()
	defer teardown()

	var ms *MetricStoreSecret
	var s SummaryMetric
	var err error
	var value float64 = 30.0

	ms, err = NewMetricStoreSecret(new_app, kd)
	assert.NoError(t, err)

	// new app; it, version, and metric get created
	err = ms.AddMetric(metric, version, value)
	assert.NoError(t, err)
	s, err = ms.GetSummaryMetric(metric, version)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), s.Count())

	// get new ms using existing app name
	ms, err = NewMetricStoreSecret(app, kd)
	assert.NoError(t, err)

	// new version; it and metric get created
	err = ms.AddMetric(metric, new_version, value)
	assert.NoError(t, err)
	s, err = ms.GetSummaryMetric(metric, new_version)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), s.Count())
	assert.Equal(t, value, s.Min())

	// new metric; is created
	err = ms.AddMetric(new_metric, version, value)
	assert.NoError(t, err)
	s, err = ms.GetSummaryMetric(new_metric, "v1")
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), s.Count())
	assert.Equal(t, value, s.Min())

	// existing metric; now see increase in count
	err = ms.AddMetric(metric, version, value)
	assert.NoError(t, err)
	s, err = ms.GetSummaryMetric(metric, version)
	assert.NoError(t, err)
	assert.Equal(t, uint32(2), s.Count())
	assert.Equal(t, float64(75), s.Sum())
	assert.Equal(t, float64(30), s.Min())
}

func TestRecordEvent(t *testing.T) {
	kd := setup()
	defer teardown()

	var ms *MetricStoreSecret
	var c MetricStoreSecretCache
	var err error

	ms, err = NewMetricStoreSecret(new_app, kd)
	assert.NoError(t, err)

	// new app; it and version are created
	err = ms.RecordEvent(VersionNewEvent, version)
	assert.NoError(t, err)
	c, err = ms.Read(version, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(c.versionData.History))

	// get new ms using previously existing app name
	ms, err = NewMetricStoreSecret(app, kd)
	assert.NoError(t, err)

	// new version; is created
	err = ms.RecordEvent(VersionNewEvent, new_version)
	assert.NoError(t, err)
	c, _ = ms.Read(new_version, metric)
	assert.Equal(t, 1, len(c.versionData.History))

	// new metric; existing version updated
	err = ms.RecordEvent(VersionNewEvent, version)
	assert.NoError(t, err)
	c, _ = ms.Read(version, new_metric)
	assert.Equal(t, 4, len(c.versionData.History))
}

func TestAddTrackEvent(t *testing.T) {
	kd := setup()
	defer teardown()

	var ms *MetricStoreSecret
	var c MetricStoreSecretCache
	var err error

	ms, err = NewMetricStoreSecret(new_app, kd)
	assert.NoError(t, err)

	// new app; it and version are created
	err = ms.AddTrackEvent(VersionMapTrackEvent, version, "track")
	assert.NoError(t, err)
	c, err = ms.Read(version, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(c.versionData.History))

}

func TestRead(t *testing.T) {
	kd := setup()
	defer teardown()

	ms, err := NewMetricStoreSecret(app, kd)
	assert.NoError(t, err)

	// existing app, version, metric
	c, err := ms.Read(version, metric)
	assert.NoError(t, err)
	assert.Equal(t, c.metricName, metric)
	assert.Equal(t, uint32(1), c.metricData.Count())

	// existing app, version, not metric
	c, err = ms.Read(version, new_metric)
	assert.Error(t, err) // metric not found: notpresent
	assert.Equal(t, uint32(0), c.metricData.Count())
	assert.Equal(t, math.MaxFloat64, c.metricData.Min())

	// existing app, no version
	c, err = ms.Read(new_version, metric)
	assert.Error(t, err) // no version data for v
	assert.Equal(t, uint32(0), c.metricData.Count())
	assert.Equal(t, math.MaxFloat64, c.metricData.Min())

	// no app
	ms, err = NewMetricStoreSecret("nosecret", kd)
	assert.NoError(t, err)

	c, err = ms.Read(version, metric)
	assert.ErrorContains(t, err, "no secret for application") // unable to read
	assert.Equal(t, uint32(0), c.metricData.Count())
	assert.Equal(t, math.MaxFloat64, c.metricData.Min())
}

func TestWrite(t *testing.T) {
}
