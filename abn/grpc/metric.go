package grpc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"math"
	"strings"
	"time"
	"unsafe"

	"github.com/iter8-tools/iter8/base/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type SummaryMetric struct {
	Count               uint32
	Sum                 float64
	Minimum             float64
	Maximum             float64
	Sumsquares          float64
	LastUpdateTimestamp [20]byte
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
	secret, err := store.client.CoreV1().Secrets(store.Namespace).Get(context.Background(), store.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("Unable to get secret")
		return nil, err
	}

	var metric SummaryMetric
	key := name + "." + version
	v, ok := secret.Data[key]
	if !ok {
		// no entry for this summary metric; create one
		metric = SummaryMetric{
			Count:      0,
			Sum:        0,
			Minimum:    math.MaxFloat64,
			Maximum:    math.SmallestNonzeroFloat64,
			Sumsquares: 0,
		}
		metric.SetLastUpdateTimestamp()
	} else {
		sDec := make([]byte, unsafe.Sizeof(metric))
		_, err = base64.StdEncoding.Decode(sDec, []byte(v))
		if err != nil {
			log.Logger.WithError(err).Error("unable to decode metric from secret")
			return nil, err
		}
		oBuf := bytes.NewReader(sDec)
		err = binary.Read(oBuf, binary.LittleEndian, &metric)
		if err != nil {
			log.Logger.WithError(err).Error("unable to read metric from secret")
			return nil, err
		}
	}
	log.Logger.Trace("MetricStoreSecret.GetSummaryMetric returns")
	log.Logger.Tracef("  count = %d, sum = %f\n", metric.Count, metric.Sum)
	return &metric, nil
}

func (metric *SummaryMetric) Add(value float64) {
	metric.Count += 1
	metric.Sum += value
	if value > metric.Maximum {
		metric.Maximum = value
	}
	if value < metric.Minimum {
		metric.Minimum = value
	}
	metric.Sumsquares += value * value
	metric.SetLastUpdateTimestamp()
}

func (store *MetricStoreSecret) Write(name string, version string, metric *SummaryMetric) error {
	log.Logger.Tracef("MetricStoreSecret.Write(%s, %s, %v)\n", name, version, metric)
	log.Logger.Tracef("  count = %d, sum = %f\n", metric.Count, metric.Sum)
	secret, err := store.client.CoreV1().Secrets(store.Namespace).Get(context.Background(), store.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("unable to read metrics store")
		return err
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, metric)
	if err != nil {
		log.Logger.WithError(err).Warn("unable to write metric")
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	secret.Data[name+"."+version] = []byte(encoded)

	_, err = store.client.CoreV1().Secrets(store.Namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		log.Logger.WithError(err).Warn("unable to update metrics store")
	}

	return nil

}

func (metric *SummaryMetric) SetLastUpdateTimestamp() {
	now := []byte(time.Now().UTC().Format(time.RFC3339))
	for i, b := range now {
		metric.LastUpdateTimestamp[i] = b
	}

}

func parseNamespacedName(namespacedName string) (string, string) {
	names := strings.Split(namespacedName, "/")
	return names[0], names[1]
}
