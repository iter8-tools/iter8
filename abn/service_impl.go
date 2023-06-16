package abn

// lookup.go -(internal) implementation of gRPC Lookup method

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/maphash"
	"strconv"
	"strings"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/base/summarymetrics"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

//
//
//

var versionHasher maphash.Hash

// lookupInternal is detailed implementation of gRPC method Lookup
// application is a namespacedname, "namespace/name"
func lookupInternal(application string, user string) (*controllers.Routemap, *int, error) {
	// if user is not provided, fail
	if user == "" {
		return nil, nil, errors.New("no user session provided")
	}

	// check that we have a record of the application
	if application == "" {
		return nil, nil, fmt.Errorf("application %s not found", application)
	}

	ns, name := splitApplicationKey(application)
	s := controllers.AllRoutemaps.GetRoutemapFromNamespaceName(ns, name)
	if s == nil {
		return nil, nil, fmt.Errorf("routemap not found for application %s", application)
	}

	track := rendezvousGet(s, user)
	if track == nil {
		return nil, nil, fmt.Errorf("no versions in routemap for application %s", application)
	}

	return s, track, nil
}

// rendezvousGet is an implementation of rendezvous hashing (cf. https://en.wikipedia.org/wiki/Rendezvous_hashing)
// It returns a consistent track for a given application and user combination.
// The track is chosen uniformly at random from among the current set of tracks
// associated with an application.
// We want to always return the same track for the same user so long as the
// application remains unchanged -- there are no change in the set of versions
// and no change to the track mapping.
// We select the version, user pair with the largest hash value ("score").
// Inspired by https://github.com/tysonmote/rendezvous/blob/master/rendezvous.go
func rendezvousGet(s *controllers.Routemap, user string) *int {
	// current maximimum score as computed by the hash function
	var maxScore uint64
	// maxTrack is the track with the current maximum score
	var maxTrack int

	// no versions
	if len(s.Versions) == 0 {
		return nil
	}

	for track := range s.Versions {
		score := hash(fmt.Sprintf("%d", track), user)
		log.Logger.Debugf("hash(%d,%s) --> %d  --  %d", track, user, score, maxScore)
		if score >= maxScore {
			maxScore = score
			maxTrack = track
		}
	}
	return &maxTrack
}

// hash computes the score for a version, user combination
func hash(track, user string) uint64 {
	versionHasher.Reset()
	_, _ = versionHasher.WriteString(user)
	_, _ = versionHasher.WriteString(track)
	return versionHasher.Sum64()
}

// nameFromKey returns the name from a key of the form "namespace/name"
func nameFromKey(applicationKey string) string {
	_, n := splitApplicationKey(applicationKey)
	return n
}

// namespaceFromKey returns the namespace from a key of the form "namespace/name"
func namespaceFromKey(applicationKey string) string {
	ns, _ := splitApplicationKey(applicationKey)
	return ns
}

// splitApplicationKey is a utility function that returns the name and namespace from a key of the form "namespace/name"
func splitApplicationKey(applicationKey string) (string, string) {
	var name, namespace string
	names := strings.Split(applicationKey, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return namespace, name
}

//
//
//

// writeMetricInternal is detailed implementation of gRPC method WriteMetric
func writeMetricInternal(application, user, metric, valueStr string, client k8sclient.Interface) error {
	log.Logger.Tracef("writeMetricInternal called for application, user: %s, %s", application, user)
	defer log.Logger.Trace("writeMetricInternal completed")

	s, track, err := lookupInternal(application, user)
	if err != nil || track == nil {
		return err
	}
	log.Logger.Debugf("lookup(%s,%s) -> %d", application, user, *track)

	s.Lock()
	defer s.Unlock()

	if s.Versions[*track].Metrics == nil {
		s.Versions[*track].Metrics = map[string]*summarymetrics.SummaryMetric{}
	}
	m, ok := s.Versions[*track].Metrics[metric]
	if !ok {
		m = summarymetrics.EmptySummaryMetric()
		s.Versions[*track].Metrics[metric] = m
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Logger.Warn("Unable to parse metric value ", valueStr)
		return err
	}
	m.Add(value)

	// persist updated metric
	legacyApp := routemapToLegacyApplication(s)
	return write(client, legacyApp)
}

func routemapToLegacyApplication(s *controllers.Routemap) abnapp.LegacyApplication {

	tracks := make(abnapp.LegacyTracks, len(s.Versions))
	versions := make(abnapp.LegacyVersions, len(s.Versions))
	for t, v := range s.Versions {
		asStr := fmt.Sprintf("%d", t)
		tracks[asStr] = asStr
		versions[asStr] = &abnapp.LegacyVersion{
			Metrics: v.Metrics,
		}
	}

	a := abnapp.LegacyApplication{
		Name:     s.Namespace + "/" + s.Name,
		Tracks:   tracks,
		Versions: versions,
	}
	return a
}

const secretKey string = "application.yaml"

func write(client k8sclient.Interface, a abnapp.LegacyApplication) error {
	var secret *corev1.Secret

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
	secret, err = client.GetSecret(secretNamespace, secretName)
	if err != nil {
		log.Logger.Error("cannot get secret; no metrics can be recorded: ", err)
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
	_, err = client.UpdateSecret(secret)
	if err != nil {
		log.Logger.WithError(err).Warn("unable to persist app data")
		return err
	}

	return nil
}

//
//
//

// getApplicationDataInternal is detailed implementation of gRPC method GetApplicationData
func getApplicationDataInternal(application string) (string, error) {

	namespace, name := splitApplicationKey(application)
	s := controllers.AllRoutemaps.GetRoutemapFromNamespaceName(namespace, name)
	if s == nil {
		return "", fmt.Errorf("routemap not found for application %s", application)
	}

	legacyApp := routemapToLegacyApplication(s)

	jsonBytes, err := json.Marshal(legacyApp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
