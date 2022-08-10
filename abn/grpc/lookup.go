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

	keyBytes := []byte(key)
	for t, v := range watcher.Apps[name].Tracks {
		vBytes := []byte(v)
		score := hash(vBytes, keyBytes)
		log.Logger.Debugf("  track %s (version %s): %d", t, v, score)
		if score > maxScore || (score == maxScore && bytes.Compare(vBytes, maxNode) < 0) {
			maxScore = score
			maxNode = vBytes
		}
	}
	return string(maxNode)
}

var hasher = crc32.New(crc32.MakeTable(crc32.Castagnoli))

func hash(node, key []byte) uint32 {
	hasher.Reset()
	hasher.Write(key)
	hasher.Write(node)
	return hasher.Sum32()
}

func Lookup(name string, user string) (*watcher.Version, error) {
	// if user is not provided, use a random string
	if user == "" {
		return nil, errors.New("no user session provided")
	}

	// get app from name, fail if not present
	app, ok := watcher.Apps[name]
	if !ok {
		return nil, errors.New("no versions found for application")
	}

	// use rendezvous hash to get version for user, fail if not present
	v := rendezvousGet(name, user)
	version, ok := app.Versions[v]
	if !ok {
		return nil, errors.New("can't find version")
	}

	log.Logger.Debugf("lookup.Lookup :: version = %s, track = %s, ready = %t", version.Name, version.Track, version.Ready)
	return &version, nil
}
