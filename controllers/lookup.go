package controllers

// lookup.go -(internal) implementation of gRPC Lookup method

import (
	"errors"
	"fmt"
	"hash/maphash"
	"strings"

	"github.com/iter8-tools/iter8/base/log"
)

var versionHasher maphash.Hash

// lookupInternal is detailed implementation of gRPC method Lookup
// application is a namespacedname, "namespace/name"
func lookupInternal(application string, user string) (*routemap, *int, error) {
	// if user is not provided, fail
	if user == "" {
		return nil, nil, errors.New("no user session provided")
	}

	// check that we have a record of the application
	if application == "" {
		return nil, nil, fmt.Errorf("application %s not found", application)
	}

	ns, name := splitApplicationKey(application)
	s := allRoutemaps.getRoutemapFromNamespaceName(ns, name)
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
func rendezvousGet(s *routemap, user string) *int {
	// current maximimum score as computed by the hash function
	var maxScore uint64
	// maxTrack is the track with the current maximum score
	var maxTrack int

	// no versions
	if len(s.Versions) == 0 {
		return nil
	}

	for track, _ := range s.Versions {
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
