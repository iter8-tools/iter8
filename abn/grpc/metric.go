package grpc

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type VersionMetric struct {
	EncodedSummaryMetric string
	LastUpdateTimestamp  string
}

type SummaryMetric struct {
	Count               uint32
	Sum                 float64
	Minimum             float64
	Maximum             float64
	Sumsquares          float64
	LastUpdateTimestamp time.Time
}
type EncodedSummaryMetric []float64

func EncodeMetric(metric *SummaryMetric) EncodedSummaryMetric {
	log.Logger.Debugf("EncodeMetric called with <%s>\n", metric.toString())
	encoded := EncodedSummaryMetric{
		float64(metric.Count),
		metric.Sum,
		metric.Minimum,
		metric.Maximum,
		metric.Sumsquares,
		float64(metric.LastUpdateTimestamp.Unix()),
	}
	log.Logger.Debug(encoded)
	return encoded
}

func DecodeMetric(encoded EncodedSummaryMetric) *SummaryMetric {
	log.Logger.Debugf("DecodeMetric(%v)\n", encoded)
	return &SummaryMetric{
		Count:               uint32(math.Round(encoded[0])),
		Sum:                 encoded[1],
		Minimum:             encoded[2],
		Maximum:             encoded[3],
		Sumsquares:          encoded[4],
		LastUpdateTimestamp: time.Unix(int64(math.Round(encoded[5])), 0),
	}
}

type VersionData struct {
	Metrics        map[string]EncodedSummaryMetric
	LastUpdateTime time.Time
}

type MetricStore interface {
	GetSummaryMetic(string, string) (*SummaryMetric, error)
	Write(*SummaryMetric) error
}

type MetricStoreSecret struct {
	// name of secret
	Name string
	// namepsace of secret
	Namespace string
	client    kubernetes.Interface
}

func NewMetricStoreSecret(name string, client *kubernetes.Clientset) *MetricStoreSecret {
	ns, nm := parseNamespacedName(name)

	return &MetricStoreSecret{
		Name:      nm,
		Namespace: ns,
		client:    client,
	}
}

func (store *MetricStoreSecret) GetSummaryMetric(name string, version string) (*SummaryMetric, error) {
	// read secret from cluster
	secret, err := store.client.CoreV1().Secrets(store.Namespace).Get(context.Background(), store.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("unable to read metrics store")
		return nil, err
	}

	var versionData = VersionData{}

	// read (encoded) version data from secret data
	rawVersionData, ok := secret.Data[version]
	if ok {
		// convert yaml string to go object (ie, map[string]VersionMetric)
		err = yaml.Unmarshal(rawVersionData, &versionData)
		if err != nil {
			log.Logger.WithError(err).Error("unable to parse versiondata")
			return nil, err
		}
	} // else {} // no secret data so vms is the empty map defined above

	// get summary metric from versionmetric
	var metric *SummaryMetric
	encodedMetric, ok := versionData.Metrics[name]
	if !ok {
		// no entry for this version; create one
		metric = &SummaryMetric{
			Count:      0,
			Sum:        0,
			Minimum:    math.MaxFloat64,
			Maximum:    math.SmallestNonzeroFloat64,
			Sumsquares: 0,
			// LastUpdateTimestamp: time.Now(),
		}
	} else {
		metric = DecodeMetric(encodedMetric)
	}

	log.Logger.Tracef("MetricStoreSecret.GetSummaryMetric returns <%s>\n", metric.toString())
	return metric, nil
}

func (metric *SummaryMetric) Add(value float64) {
	log.Logger.Tracef("Add() called with: <%s>", metric.toString())
	metric.Count += 1
	metric.Sum += value
	if value > metric.Maximum {
		metric.Maximum = value
	}
	if value < metric.Minimum {
		metric.Minimum = value
	}
	metric.Sumsquares += value * value
	metric.LastUpdateTimestamp = time.Now()
	log.Logger.Tracef("after Add(): <%s>", metric.toString())
}

func (metric *SummaryMetric) toString() string {
	return fmt.Sprintf("%s: [%d] %f < %f, (%f, %f)",
		metric.LastUpdateTimestamp,
		metric.Count,
		metric.Minimum,
		metric.Maximum,
		metric.Sum,
		metric.Sumsquares,
	)
}

func (store *MetricStoreSecret) Write(name string, version string, metric *SummaryMetric) error {
	log.Logger.Tracef("MetricStoreSecret.Write(%s, %s, <%s>)\n", name, version, metric.toString())

	// encode summary metric
	encodedMetric := EncodeMetric(metric)

	// read secret from cluster
	secret, err := store.client.CoreV1().Secrets(store.Namespace).Get(context.Background(), store.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("unable to read metrics store")
		return err
	}

	var versionData = VersionData{
		Metrics: map[string]EncodedSummaryMetric{},
	}

	// read (encoded) version data from secret data
	rawVersionData, ok := secret.Data[version]
	if ok {
		// convert yaml string to go object (ie, map[string]VersionMetric)
		err = yaml.Unmarshal(rawVersionData, &versionData)
		if err != nil {
			log.Logger.WithError(err).Error("unable to parse versiondata")
			return err
		}
	} // else {} // no secret data so vms is the empty map defined above

	versionData.Metrics[name] = encodedMetric
	versionData.LastUpdateTime = metric.LastUpdateTimestamp

	// encode vms to yaml string
	encodedVersionData, err := yaml.Marshal(versionData)
	if err != nil {
		log.Logger.WithError(err).Warn("unable to encode versionmetrics")
		return err
	}

	// put entry in secret data (base64 encode it first)
	secret.Data[version] = encodedVersionData // []byte(base64.StdEncoding.EncodeToString(encodedVMs))

	// write secret to cluster
	_, err = store.client.CoreV1().Secrets(store.Namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("unable to update metrics store")
	}

	return nil
}

func parseNamespacedName(namespacedName string) (string, string) {
	names := strings.Split(namespacedName, "/")
	return names[0], names[1]
}
