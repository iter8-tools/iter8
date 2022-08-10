package appsummary

// methods to read/write application data (to a Kubernetes secret). In this context, application data
// is, for each version, (a) a list of summary metrics and their values, and (b) a history of the version
// including such events as discovery, readiness, and mapping to track

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/yaml"
)

const (
	VERSIONS_DATA = "versionData.yaml"
)

type ApplicationSummary struct {
	Name string
	Data ApplicationData
}

type ApplicationData map[string]VersionData

type Version struct {
	Name string
	Data VersionData
}

type VersionData struct {
	Metrics             map[string]SummaryMetric `json:"metrics" yaml:"metrics"`
	LastUpdateTimestamp time.Time                `json:"lastUpdateTimestamp" yaml:"lastUpdateTimestamp"`
	History             []VersionEvent           `json:"history" yaml:"history"`
}

type VersionEvent struct {
	Timestamp time.Time        `json:"tm" yaml:"tm"`
	Event     VersionEventType `json:"ev" yaml:"ev"`
	Track     string           `json:"trk,omitempty" yaml:"trk,omitempty"`
}

type VersionEventType string

const (
	VersionNewEvent           VersionEventType = "new"
	VersionReadyEvent         VersionEventType = "ready"
	VersionNoLongerReadyEvent VersionEventType = "notready"
	VersionMapTrackEvent      VersionEventType = "track"
	VersionUnmapTrackEvent    VersionEventType = "untrack"
)

type MetricDriver struct {
	// client is kubernetes client used to read/write the secret
	Client kubernetes.Interface
}

func (driver *MetricDriver) ReadApplicationSummary(application string) (ApplicationSummary, error) {
	summary := ApplicationSummary{
		Name: application,
		Data: ApplicationData{},
	}

	name, namespace := AppNameToSecretNameNamespace(application)
	// read secret from cluster; extract appData
	secret, err := driver.Client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Logger.Debugf("ReadApplication no secret returning %+v", summary)
		return summary, err
	}

	// read data from secret (is a yaml file)
	rawData, ok := secret.Data[VERSIONS_DATA]
	if !ok {
		log.Logger.Debugf("ReadApplication key missing returning %+v", summary)
		return summary, errors.New("secret does not contain expected key: " + VERSIONS_DATA)
	}

	var appData ApplicationData
	err = yaml.Unmarshal(rawData, &appData)
	if err != nil {
		log.Logger.Debugf("ReadApplication unmarshal failure returning %+v", summary)
		return summary, nil
	}

	summary.Data = appData
	log.Logger.Debugf("ReadApplication returning %+v", summary)
	return summary, nil
}

func (driver *MetricDriver) WriteApplicationSummary(summary ApplicationSummary) error {
	var secret *corev1.Secret

	// marshal to byte array
	rawData, err := yaml.Marshal(summary.Data)
	if err != nil {
		return err
	}

	name, namespace := AppNameToSecretNameNamespace(summary.Name)

	exists := true
	secret, err = driver.Client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		exists = false
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data:       map[string][]byte{},
			StringData: map[string]string{},
		}
		log.Logger.Infof("creating secret %#v", secret)
	}

	secret.Data[VERSIONS_DATA] = rawData
	if secret.StringData != nil {
		secret.StringData[VERSIONS_DATA] = string(rawData)
	}

	// create or update the secret
	if exists {
		_, err = driver.Client.CoreV1().Secrets(namespace).Update(
			context.Background(),
			secret,
			metav1.UpdateOptions{},
		)
	} else {
		_, err = driver.Client.CoreV1().Secrets(namespace).Create(
			context.Background(),
			secret,
			metav1.CreateOptions{},
		)
	}
	if err != nil {
		log.Logger.WithError(err).Warn("unable to update application summary")
	}

	return err
}

func (a ApplicationSummary) GetVersion(version string) (Version, error) {
	v := Version{
		Name: version,
		Data: VersionData{
			Metrics:             map[string]SummaryMetric{},
			History:             []VersionEvent{},
			LastUpdateTimestamp: time.Now(),
		},
	}

	versionData, ok := a.Data[version]
	if !ok {
		log.Logger.Debugf("GetVersion no data found; returning %+v", v)
		return v, errors.New("no data found for version " + version)
	}

	v.Data = versionData
	log.Logger.Debugf("GetVersion returning %+v", v)
	return v, nil
}

func (v Version) GetMetric(metric string) (SummaryMetric, error) {
	m, ok := v.Data.Metrics[metric]
	if !ok {
		m = EmptySummaryMetric()
		log.Logger.Debugf("GetMetric not found;returning %+v", m)
		return m, errors.New("metric not found " + metric)
	}

	log.Logger.Debugf("GetMetric returning %+v", m)
	return m, nil
}

func (driver *MetricDriver) AddMetric(application string, version string, metric string, value float64) error {
	a, _ := driver.ReadApplicationSummary(application)
	v, _ := a.GetVersion(version)
	m, _ := v.GetMetric(metric)

	m = m.AddTo(value)
	v.Data.LastUpdateTimestamp = time.Now()

	v.Data.Metrics[metric] = m
	a.Data[version] = v.Data
	err := driver.WriteApplicationSummary(a)

	return err
}

func (driver *MetricDriver) RecordEvent(application string, version string, event VersionEventType, track ...string) error {
	a, _ := driver.ReadApplicationSummary(application)
	v, _ := a.GetVersion(version)

	vEvent := VersionEvent{
		Event:     event,
		Timestamp: time.Now(),
	}
	if event == VersionMapTrackEvent {
		if len(track) != 1 {
			return errors.New("map track event requires one track")
		}
		vEvent.Track = track[0]
	}
	v.Data.History = append(v.Data.History, vEvent)
	v.Data.LastUpdateTimestamp = time.Now()

	a.Data[version] = v.Data
	err := driver.WriteApplicationSummary(a)
	return err
}

// AppNameToSecretNameNamespace converts application name to secret name/namespace
func AppNameToSecretNameNamespace(application string) (name string, namespace string) {
	names := strings.Split(application, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return name, namespace
}
