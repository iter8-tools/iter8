package abn

// lookup.go -(internal) implementation of gRPC Lookup method

import (
	"errors"
	"fmt"
	"hash/maphash"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers"
)

var allRoutemaps controllers.AllRouteMapsInterface = &controllers.DefaultRoutemaps{}

// versionHasher is hash used for selecting versions
var versionHasher maphash.Hash

const invalidVersion int = -1

// lookupInternal is detailed implementation of gRPC method Lookup
// application is a of the form "namespace/name"
func lookupInternal(application string, user string) (controllers.RoutemapInterface, int, error) {
	// if user is not provided, fail
	if user == "" {
		return nil, invalidVersion, errors.New("no user session provided")
	}

	// check that we have a record of the application
	if application == "/" {
		return nil, invalidVersion, errors.New("no application provided")
	}

	ns, name := util.SplitApplication(application)
	s := allRoutemaps.GetAllRoutemaps().GetRoutemapFromNamespaceName(ns, name)
	if s == nil || reflect.ValueOf(s).IsNil() {
		return nil, invalidVersion, fmt.Errorf("routemap not found for application %s", ns+"/"+name)
	}

	versionNumber := rendezvousGet(s, user)
	if versionNumber == invalidVersion {
		return nil, invalidVersion, fmt.Errorf("no versions in routemap for application %s", ns+"/"+name)
	}

	// record user; ignore error if any; this is best effort
	if MetricsClient == nil {
		return nil, invalidVersion, fmt.Errorf("no metrics client")
	}
	_ = MetricsClient.SetUser(application, versionNumber, *s.GetVersions()[versionNumber].GetSignature(), user)

	return s, versionNumber, nil
}

// rendezvousGet is an implementation of rendezvous hashing (cf. https://en.wikipedia.org/wiki/Rendezvous_hashing)
// It returns a consistent versionNumber (index) for a given application and user combination.
// The version number is chosen uniformly at random from among the current set of versions
// associated with an application.
// We want to always return the same version number for the same user so long as the
// application remains unchanged -- there are no change in the set of versions
// and no change to the version number mapping.
// We select the version, user pair with the largest hash value ("score").
// Inspired by https://github.com/tysonmote/rendezvous/blob/master/rendezvous.go
func rendezvousGet(s controllers.RoutemapInterface, user string) int {
	// current maximimum score as computed by the hash function
	var maxScore float64
	// maxVersionNumber is the version index with the current maximum score
	var maxVersionNumber int

	// no versions
	processedVersions := 0

	s.RLock()
	defer s.RUnlock()

	sumW := uint32(0)
	for versionNumber := range s.GetVersions() {
		sumW += s.Weights()[versionNumber]
	}

	for versionNumber, version := range s.GetVersions() {
		w := s.Weights()[versionNumber]
		if w == 0 {
			continue
		}
		wFactor := float64(w) / float64(sumW)
		h := hash(fmt.Sprintf("%d", versionNumber), *version.GetSignature(), user)
		score := wFactor * float64(h)
		log.Logger.Debugf("hash(%d,%s) --> %f  --  %f", versionNumber, user, score, maxScore)
		if score >= maxScore {
			maxScore = score
			maxVersionNumber = versionNumber
		}
		processedVersions++
	}

	// if no versions (available; ie, non-zero weight)
	if processedVersions == 0 {
		return invalidVersion
	}
	return maxVersionNumber
}

// hash computes the score for a version, user combination
func hash(version, signature, user string) uint64 {
	versionHasher.Reset()
	_, _ = versionHasher.WriteString(user)
	_, _ = versionHasher.WriteString(signature)
	_, _ = versionHasher.WriteString(version)
	return versionHasher.Sum64()
}

// writeMetricInternal is detailed implementation of gRPC method WriteMetric
func writeMetricInternal(application, user, metric, valueStr string) error {
	log.Logger.Tracef("writeMetricInternal called for application, user: %s, %s", application, user)
	defer log.Logger.Trace("writeMetricInternal completed")

	s, versionNumber, err := lookupInternal(application, user)
	if err != nil || versionNumber == invalidVersion {
		log.Logger.Warnf("lookupInternal failed for application=%s, user=%s", application, user)
		return err
	}
	log.Logger.Debugf("lookupInternal(%s,%s) -> %d", application, user, versionNumber)

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Logger.Warn("Unable to parse metric value ", valueStr)
		return err
	}

	v := s.GetVersions()[versionNumber]
	transaction := uuid.NewString()

	if MetricsClient == nil {
		return fmt.Errorf("no metrics client")
	}
	err = MetricsClient.SetMetric(
		s.GetNamespace()+"/"+s.GetName(), versionNumber, *v.GetSignature(),
		metric, user, transaction,
		value)

	if err != nil {
		log.Logger.Warnf("Unable to set metric %s for application=%s, user=%s", metric, application, metric)
	}

	return nil
}
