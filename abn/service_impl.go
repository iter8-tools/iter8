package abn

// lookup.go -(internal) implementation of gRPC Lookup method

import (
	"errors"
	"fmt"
	"hash/maphash"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers"
)

var allRoutemaps controllers.AllRouteMapsInterface = &controllers.DefaultRoutemaps{}

// versionHasher is hash used for selecting versions
var versionHasher maphash.Hash

// lookupInternal is detailed implementation of gRPC method Lookup
// application is a of the form "namespace/name"
func lookupInternal(application string, user string) (controllers.RoutemapInterface, *int, error) {
	// func lookupInternal(application string, user string, routemaps controllers.RoutemapsInterface) (controllers.RoutemapInterface, *int, error) {
	// if user is not provided, fail
	if user == "" {
		return nil, nil, errors.New("no user session provided")
	}

	// check that we have a record of the application
	if application == "/" {
		return nil, nil, errors.New("no application provided")
	}

	ns, name := splitApplicationKey(application)
	s := allRoutemaps.GetAllRoutemaps().GetRoutemapFromNamespaceName(ns, name)
	if s == nil || reflect.ValueOf(s).IsNil() {
		return nil, nil, fmt.Errorf("routemap not found for application %s", ns+"/"+name)
	}

	track := rendezvousGet(s, user)
	if track == nil {
		return nil, nil, fmt.Errorf("no versions in routemap for application %s", ns+"/"+name)
	}

	// record user; ignore error if any; this is best effort
	_ = MetricsClient.SetUser(application, *track, *s.GetVersions()[*track].GetSignature(), user)

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
func rendezvousGet(s controllers.RoutemapInterface, user string) *int {
	// current maximimum score as computed by the hash function
	var maxScore uint64
	// maxTrack is the track with the current maximum score
	var maxTrack int

	// no versions
	processedVersions := 0

	s.RLock()
	defer s.RUnlock()

	for track, version := range s.GetVersions() {
		if s.Weights()[track] == 0 {
			continue
		}
		score := hash(fmt.Sprintf("%d", track), *version.GetSignature(), user)
		log.Logger.Debugf("hash(%d,%s) --> %d  --  %d", track, user, score, maxScore)
		if score >= maxScore {
			maxScore = score
			maxTrack = track
		}
		processedVersions++
	}

	// if no versions (available; ie, non-zero weight)
	if processedVersions == 0 {
		return nil
	}
	return &maxTrack
}

// hash computes the score for a version, user combination
func hash(track, signature, user string) uint64 {
	versionHasher.Reset()
	_, _ = versionHasher.WriteString(user)
	_, _ = versionHasher.WriteString(signature)
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

	v := s.GetVersions()[*track]
	transaction := uuid.NewString()

	err = MetricsClient.SetMetric(
		s.GetNamespace()+"/"+s.GetName(), *track, *v.GetSignature(),
		metric, user, transaction,
		value)

	if err != nil {
		log.Logger.Warnf("Unable to set metric %s for application=%s, user=%s", metric, application, metric)
	}

	return nil
}
