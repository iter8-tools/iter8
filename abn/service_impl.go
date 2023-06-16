package abn

// lookup.go -(internal) implementation of gRPC Lookup method

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/maphash"
	"strconv"
	"strings"

	"github.com/google/uuid"
	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/base/summarymetrics"
	"github.com/iter8-tools/iter8/controllers"
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
func writeMetricInternal(application, user, metric, valueStr string) error {
	log.Logger.Tracef("writeMetricInternal called for application, user: %s, %s", application, user)
	defer log.Logger.Trace("writeMetricInternal completed")

	s, track, err := lookupInternal(application, user)
	if err != nil || track == nil {
		log.Logger.Warnf("lookupInternal failed for application=%s, user=%s", application, user)
		return err
	}
	log.Logger.Debugf("lookupInternal(%s,%s) -> %d", application, user, *track)

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Logger.Warn("Unable to parse metric value ", valueStr)
		return err
	}

	v := s.Versions[*track]
	transaction := uuid.NewString()

	err = metricsClient.SetMetric(
		s.Namespace+"/"+s.Name, *track, *v.Signature,
		metric, user, transaction,
		value)

	if err != nil {
		log.Logger.Warnf("Unable to set metric %s for application=%s, user=%s", metric, application, metric)
	}

	return nil
}

func toLegacyApplication(s *controllers.Routemap) *abnapp.LegacyApplication {
	name := s.Namespace + "/" + s.Name
	tracks := make(abnapp.LegacyTracks, len(s.Versions))
	versions := make(abnapp.LegacyVersions, len(s.Versions))
	for t, v := range s.Versions {
		asStr := fmt.Sprintf("%d", t)
		tracks[asStr] = asStr

		vms, err := metricsClient.GetSummaryMetrics(name, t, *v.Signature)
		if err != nil {
			return nil
		}
		metrics := make(map[string]*summarymetrics.SummaryMetric, len(vms.MetricSummaries))
		for metric, summary := range vms.MetricSummaries {
			metrics[metric] = &summarymetrics.SummaryMetric{
				float64(summary.SummaryOverTransactions.Count),
				float64(summary.SummaryOverTransactions.Count) * summary.SummaryOverTransactions.Mean,
				summary.SummaryOverTransactions.Min,
				summary.SummaryOverTransactions.Max,
				summary.SummaryOverTransactions.StdDev, // should be sum of squares
			}
		}

		versions[asStr] = &abnapp.LegacyVersion{
			Metrics: metrics,
		}
	}

	a := &abnapp.LegacyApplication{
		Name:     name,
		Tracks:   tracks,
		Versions: versions,
	}
	return a
}

// getApplicationDataInternal is detailed implementation of gRPC method GetApplicationData
func getApplicationDataInternal(application string) (string, error) {

	namespace, name := splitApplicationKey(application)
	s := controllers.AllRoutemaps.GetRoutemapFromNamespaceName(namespace, name)
	if s == nil {
		return "", fmt.Errorf("routemap not found for application %s", application)
	}

	legacyApp := toLegacyApplication(s)

	jsonBytes, err := json.Marshal(legacyApp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
