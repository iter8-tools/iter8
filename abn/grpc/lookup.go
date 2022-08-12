package grpc

// lookup.go -(internal) implementation of gRPC Lookup method

import (
	"bytes"
	"errors"
	"hash/crc32"

	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/base/log"
)

func rendezvousGet(name string, key string) string {
	var maxScore uint32
	var maxNode []byte
	var track string

	keyBytes := []byte(key)
	for t, v := range watcher.Applications[name].Tracks {
		vBytes := []byte(v)
		score := hash(vBytes, keyBytes)
		log.Logger.Debugf("  track %s (version %s): %d", t, v, score)
		if score > maxScore || (score == maxScore && bytes.Compare(vBytes, maxNode) < 0) {
			maxScore = score
			maxNode = vBytes
			track = t
		}
	}
	return track
}

var hasher = crc32.New(crc32.MakeTable(crc32.Castagnoli))

func hash(node, key []byte) uint32 {
	hasher.Reset()
	hasher.Write(key)
	hasher.Write(node)
	return hasher.Sum32()
}

// func Lookup(name string, user string) (*app.Version, error) {
// 	// if user is not provided, use a random string
// 	if user == "" {
// 		return nil, errors.New("no user session provided")
// 	}

// 	// get app from name, fail if not present
// 	a, ok := watcher.Applications[name]
// 	if !ok {
// 		return nil, errors.New("no versions found for application")
// 	}

// 	// use rendezvous hash to get version for user, fail if not present
// 	version := rendezvousGet(name, user)
// 	v, new := a.Versions[version]
// 	if new {
// 		return nil, errors.New("can't find version")
// 	}

// 	log.Logger.Debugf("lookup.Lookup :: version = %s, track = %v, ready = %v", version, *v.GetTrack(), v.IsReady())
// 	return &v, nil
// }

func Lookup(application string, user string) (*string, error) {
	// if user is not provided, fail
	if user == "" {
		return nil, errors.New("no user session provided")
	}

	// check that we have a record of the application
	_, ok := watcher.Applications[application]
	if !ok {
		return nil, errors.New("application not found")
	}

	// use rendezvous hash to get track for user, fail if not present
	track := rendezvousGet(application, user)
	return &track, nil
}
