package metricstore

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	VERSIONS_DATA = "versionData.yaml"
)

type ApplicationData map[string]VersionData

type VersionData struct {
	Metrics             map[string]SummaryMetric `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	LastUpdateTimestamp time.Time                `json:"lastUpdateTimestamp" yaml:"lastUpdateTimestamp"`
	History             []VersionEvent           `json:"history,omitempty" yaml:"history,omitempty"`
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

// A MetricStore implemented as a Kubernetes secret
type MetricStoreSecret struct {
	// App name
	App string
	// Name of the ecret to use
	Name string
	// Namepsace of the secret
	Namespace string
	// client is kubernetes client used to read/write the secret
	client kubernetes.Interface
}

// NewMetricStoreSecret creates a new MetricStoreSecret
func NewMetricStoreSecret(app string, client *kubernetes.Clientset) (*MetricStoreSecret, error) {
	if client == nil {
		return nil, errors.New("Kubernestes clientset required to create MetricStoreSecret")
	}

	var ns, nm string
	names := strings.Split(app, "/")
	if len(names) > 1 {
		ns, nm = names[0], names[1]
	} else {
		ns, nm = "default", names[0]
	}

	return &MetricStoreSecret{
		App:       app,
		Name:      nm,
		Namespace: ns,
		client:    client,
	}, nil
}

type MetricStoreSecretCache struct {
	secret      *v1.Secret
	appData     ApplicationData
	versionName string
	versionData VersionData
	metricName  string
	metricData  SummaryMetric
}

func (store *MetricStoreSecret) GetSummaryMetric(metric string, version string) (SummaryMetric, error) {
	cache, err := store.Read(version, metric)
	return cache.metricData, err
}

func (store *MetricStoreSecret) AddMetric(metric string, version string, value float64) error {
	cache, _ := store.Read(version, metric)

	// even if there is an error, we will try to write the value anyway
	cache.metricData = *cache.metricData.AddTo(value)

	return store.Write(cache)
}

func (store *MetricStoreSecret) RecordEvent(event VersionEventType, version string, track ...string) error {
	cache, _ := store.Read(version, "")

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

	cache.versionData.History = append(cache.versionData.History, vEvent)
	return store.Write(cache)
}

func (store *MetricStoreSecret) AddTrackEvent(event VersionEventType, version string, track string) error {
	cache, _ := store.Read(version, "")

	vEvent := VersionEvent{
		Event:     event,
		Timestamp: time.Now(),
		Track:     track,
	}

	cache.versionData.History = append(cache.versionData.History, vEvent)
	return store.Write(cache)
}

func (store *MetricStoreSecret) Read(version string, metric string) (MetricStoreSecretCache, error) {
	cache := MetricStoreSecretCache{
		appData:     ApplicationData{},
		versionName: version,
		versionData: VersionData{
			Metrics:             map[string]SummaryMetric{},
			History:             []VersionEvent{},
			LastUpdateTimestamp: time.Now(),
		},
		metricName: metric,
		metricData: SummaryMetric{
			0,                           // Count
			0,                           // Sum
			math.MaxFloat64,             // Min
			math.SmallestNonzeroFloat64, // Max
			0,                           // SumSquares
			float64(time.Now().Unix()),  // LastUpdateTimestamp
		},
	}

	// read secret from cluster; extract appData
	secret, err := store.client.CoreV1().Secrets(store.Namespace).Get(context.Background(), store.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("unable to read metric store secret")
		return cache, err
	}
	cache.secret = secret

	// read data from secret (is a yaml file)
	rawAppData, ok := secret.Data[VERSIONS_DATA]
	if !ok {
		log.Logger.Warn("unable to read data from secret")
		return cache, errors.New("expected key not found in secret")
	}

	var appData ApplicationData
	err = yaml.Unmarshal(rawAppData, &appData)
	if err != nil {
		log.Logger.WithError(err).Warn("unable to unmarshal appData from secret")
		return cache, err
	}
	if appData != nil {
		// when nil, leave the default empty map already in cache
		cache.appData = appData
	}

	var versionData VersionData
	versionData, ok = appData[version]
	if !ok {
		log.Logger.Warnf("no version data for %s", version)
		return cache, errors.New("no data found for version")
	}
	if versionData.Metrics == nil {
		versionData.Metrics = map[string]SummaryMetric{}
	}
	cache.versionData = versionData

	if metric == "" {
		return cache, nil
	}

	metricData, ok := versionData.Metrics[metric]
	if !ok {
		log.Logger.Warnf("metric not found: %s", metric)
		return cache, errors.New("no value found")
	}
	cache.metricData = metricData

	return cache, nil
}

func (store *MetricStoreSecret) Write(cache MetricStoreSecretCache) (err error) {
	if cache.secret == nil {
		// there was no secret; we fail
		return errors.New("invalid secret for metrics store")
	}
	if cache.metricName != "" {
		cache.versionData.Metrics[cache.metricName] = cache.metricData
	}
	cache.versionData.LastUpdateTimestamp = time.Now()
	cache.appData[cache.versionName] = cache.versionData

	// marshal to byte array
	rawAppData, err := yaml.Marshal(cache.appData)
	if err != nil {
		return err
	}

	// assign to secret data and update cluster
	cache.secret.Data[VERSIONS_DATA] = rawAppData
	_, err = store.client.CoreV1().Secrets(store.Namespace).Update(context.Background(), cache.secret, metav1.UpdateOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("unable to update metrics store")
	}

	return err
}
