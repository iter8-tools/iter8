package abn

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/metrics"
	"github.com/iter8-tools/iter8/storage/badgerdb"
	"github.com/stretchr/testify/assert"
)

// tests that we get the same result for the same inputs
func TestLookupInternal(t *testing.T) {
	var err error
	// set up test metrics db for recording users
	tempDirPath := t.TempDir()
	metrics.MetricsClient, err = badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
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

func TestWeights(t *testing.T) {
	var err error

	// set up test metrics db for recording users
	tempDirPath := t.TempDir()
	metrics.MetricsClient, err = badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)

	// setup: add desired routemaps to allRoutemaps
	testRM := testRoutemaps{
		allroutemaps: setupRoutemaps(t, *getWeightedTestRM("default", "test", []uint32{3, 1})),
	}
	allRoutemaps = &testRM

	tries := 100
	versionNumbers := make([]int, tries)
	for i := 0; i < tries; i++ {
		_, v, err := lookupInternal("default/test", uuid.NewString())
		assert.NoError(t, err)
		versionNumbers[i] = v
	}

	// expect 3/4 will be for version 0 (weight 3); ie, 75
	// expect 1/4 will be for version 1 (weight 1); ie, 25
	// compute number for version 1 by summing versionNumbers
	// assert less than 30 (bigger than 25)
	// there is a slight possibility of test failure

	sum := 0
	for i := 1; i < tries; i++ {
		sum += versionNumbers[i]
	}
	assert.Less(t, sum, 30)
}

func getWeightedTestRM(namespace, name string, weights []uint32) *testroutemap {
	copyWeights := make([]uint32, len(weights))
	versions := make([]testversion, len(weights))
	for i := range weights {
		copyWeights[i] = weights[i]
		versions[i] = testversion{signature: util.StringPointer(uuid.NewString())}
	}

	return &testroutemap{
		namespace:         namespace,
		name:              name,
		versions:          versions,
		normalizedWeights: copyWeights,
	}

}
