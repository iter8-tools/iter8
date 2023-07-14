package abn

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/iter8-tools/iter8/storage/badgerdb"
	"github.com/stretchr/testify/assert"
)

// tests that we get the same result for the same inputs
func TestLookupInternal(t *testing.T) {
	var err error
	// set up test metrics db for recording users
	tempDirPath := t.TempDir()
	MetricsClient, err = badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)

	// setup: add desired routemaps to allRoutemaps
	testRM := testRoutemaps{
		allroutemaps: setupRoutemaps(t, *getTestRM("default", "test")),
	}
	allRoutemaps = &testRM

	tries := 20 // needs to be big enough to find at least one problem; this is probably overkill
	// do lookup tries times
	versionNumbers := make([]int, tries)
	for i := 0; i < tries; i++ {
		_, v, err := lookupInternal("default/test", "user")
		assert.NoError(t, err)
		versionNumbers[i] = v
	}

	tr := versionNumbers[0]
	for i := 1; i < tries; i++ {
		assert.Equal(t, tr, versionNumbers[i])
	}
}
