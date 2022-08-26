package grpc

// lookup.go -(internal) implementation of gRPC Lookup method

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
)

func rendezvousGet(a *abnapp.Application, key string) string {
	var maxScore uint32
	var maxNode []byte
	var track string

	keyBytes := []byte(key)
	for t, v := range a.Tracks {
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

func Lookup(application string, user string) (*string, error) {
	// if user is not provided, fail
	if user == "" {
		return nil, errors.New("no user session provided")
	}

	// check that we have a record of the application
	a, err := abnapp.Applications.Get(application, true)
	if err != nil {
		return nil, fmt.Errorf("application not found: %s", err.Error())
	}
	if a == nil {
		return nil, errors.New("application not found")
	}

	// use rendezvous hash to get track for user, fail if not present
	track := rendezvousGet(a, user)
	return &track, nil
}
