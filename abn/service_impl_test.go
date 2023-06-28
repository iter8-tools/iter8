package abn

import (
	"os"
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
	allroutemaps := setupRoutemaps(t, *getTestRM("default", "test"))

	tries := 20 // needs to be big enough to find at least one problem; this is probably overkill
	// do lookup tries times
	tracks := make([]*int, tries)
	for i := 0; i < tries; i++ {
		_, tr, err := lookupInternal("default/test", "user", &allroutemaps)
		assert.NoError(t, err)
		tracks[i] = tr
	}

	tr := tracks[0]
	for i := 1; i < tries; i++ {
		assert.Equal(t, *tr, *tracks[i])
	}
}

func TestGetApplicationDataInternal(t *testing.T) {
	namespace, name := "default", "test"

	// setup: add desired routemaps to allRoutemaps
	allroutemaps := setupRoutemaps(t, *getTestRM(namespace, name))

	var err error
	tempDirPath := t.TempDir()
	MetricsClient, err = badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)

	// add a metric value
	err = writeMetricInternal(namespace+"/"+name, "user", "metric", "45", &allroutemaps)
	assert.NoError(t, err)

	// add a second value to metric
	err = writeMetricInternal(namespace+"/"+name, "user", "metric", "55", &allroutemaps)
	assert.NoError(t, err)

	_, tr, err := lookupInternal(namespace+"/"+name, "user", &allroutemaps)
	assert.NoError(t, err)
	assert.NotNil(t, tr)

	// get data from storage; fails because
	_, err = getApplicationDataInternal(namespace+"/"+name, &allroutemaps)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not supported")
}

func TestGetVolumeUsage(t *testing.T) {
	// GetVolumeUsage is based off of statfs which analyzes the volume, not the directory
	// Creating a temporary directory will not change anything
	path, err := os.Getwd()
	assert.NoError(t, err)

	availableBytes, totalBytes, err := GetVolumeUsage(path)
	assert.NoError(t, err)

	// The volume should have some available and total bytes
	assert.NotEqual(t, 0, availableBytes)
	assert.NotEqual(t, 0, totalBytes)
}
