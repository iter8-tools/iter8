package core

import (
	"testing"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/stretchr/testify/assert"
)

func TestLookupInternal(t *testing.T) {
	// set up by adding single application w/ multiple tracks
	abnapp.Applications.Clear()
	a := abnapp.NewApplication("default/application")
	a.Versions["v1"] = &abnapp.Version{}
	a.Versions["v2"] = &abnapp.Version{}
	a.Tracks["default"] = "v1"
	a.Tracks["candidate"] = "v2"
	abnapp.Applications.Put(a)

	tries := 20 // needs to be big enough to find at least one problem; this is probably overkill

	// do lookup tries times
	tracks := make([]*string, tries)
	for i := 0; i < tries; i++ {
		_, tr, err := lookupInternal("default/application", "user")
		assert.NoError(t, err)
		tracks[i] = tr
	}

	tr := tracks[0]
	for i := 1; i < tries; i++ {
		assert.Equal(t, *tr, *tracks[i])
	}
}
