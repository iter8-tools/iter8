package controllers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/base/summarymetrics"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// writeMetricInternal is detailed implementation of gRPC method WriteMetric
func writeMetricInternal(application, user, metric, valueStr string) error {
	log.Logger.Tracef("writeMetricInternal called for application, user: %s, %s", application, user)
	defer log.Logger.Trace("writeMetricInternal completed")

	s, track, err := lookupInternal(application, user)
	if err != nil || track == nil {
		return err
	}
	log.Logger.Debug(fmt.Sprintf("lookup(%s,%s) -> %d", application, user, *track))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	v := s.Versions[*track]
	for n, m := range v.Metrics {
		if n == metric {
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				log.Logger.Warn("Unable to parse metric value ", valueStr)
				return err
			}
			m.Add(value)
			break
		}
	}

	// persist updated metric
	legacyApp := routemapToLegacyApplication(s)
	return write(legacyApp)
}

// applicatoon is an legacy object capturing application details
type legacyApplication struct {
	// Name is of the form namespace/Name where the Name is the value of the label app.kubernetes.io/Name
	Name string `json:"name" yaml:"name"`
	// Tracks is map from application track identifier to version name
	Tracks legacyTracks `json:"tracks" yaml:"tracks"`
	// Versions maps version name to version data (a set of metrics)
	Versions legacyVersions `json:"versions" yaml:"versions"`
}

// legacyVersions is a map of the version name to a version object
type legacyVersions map[string]*legacyVersion

// legacyTracks is map of track identifiers to version names
type legacyTracks map[string]string

// legacyVersion is information about versions of an application in a Kubernetes cluster
type legacyVersion struct {
	// List of (summary) metrics for a version
	Metrics map[string]*summarymetrics.SummaryMetric `json:"metrics" yaml:"metrics"`
}

func routemapToLegacyApplication(s *routemap) legacyApplication {

	tracks := make(legacyTracks, len(s.Versions))
	versions := make(legacyVersions, len(s.Versions))
	for t, v := range s.Versions {
		asStr := fmt.Sprintf("%d", t)
		tracks[asStr] = asStr
		versions[asStr] = &legacyVersion{
			Metrics: v.Metrics,
		}
	}

	a := legacyApplication{
		Name:     s.Namespace + "/" + s.Name,
		Tracks:   tracks,
		Versions: versions,
	}

	return a
}

const secretKey string = "application.yaml"

func write(a legacyApplication) error {
	var secret *corev1.Secret

	client, err := k8sclient.New(cli.New())
	if err != nil {
		log.Logger.Error("could not obtain Kube client ... ")
		return err
	}

	// marshal to byte array
	// note that this uses a custom MarshalJSON that removes untracked
	// versions if the application data is too large
	rawData, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	secretNamespace := namespaceFromKey(a.Name)
	secretName := nameFromKey(a.Name)

	// get the current secret; it will have been created as part of install
	secret, err = client.ClientSet().CoreV1().Secrets(secretNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Logger.Error("secret does not exist; no metrics can be recorded")
		return err
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	secret.Data[secretKey] = rawData
	if secret.StringData != nil {
		secret.StringData[secretKey] = string(rawData)
	}

	// update the secret
	// TBD do we need to merge what we have?
	_, err = client.ClientSet().CoreV1().Secrets(secretNamespace).Update(
		context.Background(),
		secret,
		metav1.UpdateOptions{},
	)
	if err != nil {
		log.Logger.WithError(err).Warn("unable to persist app data")
		return err
	}

	// // update last write time for application
	// now := time.Now()
	// m.lastWriteTimes[a.Name] = &now
	return nil

}
