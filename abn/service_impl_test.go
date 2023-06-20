package abn

import (
	"os"
	"testing"

	"github.com/dgraph-io/badger/v4"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/controllers/storageclient/badgerdb"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// tests that we get the same result for the same inputs
func TestLookupInternal(t *testing.T) {
	namespace, name := "default", "test"

	// setup: add desired routemaps to allRoutemaps; first clear all routemaps
	controllers.AllRoutemaps.Clear()
	controllers.AllRoutemaps.AddRouteMap(namespace, name, &controllers.Routemap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Versions: []controllers.Version{
			{Signature: util.StringPointer("123456789")},
			{Signature: util.StringPointer("987654321")},
		},
		NormalizedWeights: []uint32{1, 1},
	})

	tries := 20 // needs to be big enough to find at least one problem; this is probably overkill
	// do lookup tries times
	tracks := make([]*int, tries)
	for i := 0; i < tries; i++ {
		_, tr, err := lookupInternal("default/test", "user")
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

	// setup
	setupRouteMaps(t, namespace, name)
	var err error
	tempDirPath := t.TempDir()
	metricsClient, err = badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)

	// add a metric value
	err = writeMetricInternal(namespace+"/"+name, "user", "metric", "45")
	assert.NoError(t, err)

	// add a second value to metric
	err = writeMetricInternal(namespace+"/"+name, "user", "metric", "55")
	assert.NoError(t, err)

	_, tr, err := lookupInternal("default/test", "user")
	assert.NoError(t, err)
	assert.NotNil(t, tr)

	// get data from secret
	app, err := getApplicationDataInternal(namespace + "/" + name)
	assert.NoError(t, err)
	assert.NotNil(t, app)

	// verify result
	if *tr == 0 {
		// assert.Equal(t, "{\"name\":\"default/test\",\"tracks\":{\"0\":\"0\",\"1\":\"1\"},\"versions\":{\"0\":{\"metrics\":{\"metric\":[2,100,45,55,5]}},\"1\":{\"metrics\":null}}}", app)
		assert.Equal(t, "{\"name\":\"default/test\",\"tracks\":{\"0\":\"0\",\"1\":\"1\"},\"versions\":{\"0\":{\"metrics\":{\"metric\":[2,100,45,55,5]}},\"1\":{\"metrics\":{\"\":[0,0,0,0,0]}}}}", app)
	} else {
		// assert.Equal(t, "{\"name\":\"default/test\",\"tracks\":{\"0\":\"0\",\"1\":\"1\"},\"versions\":{\"0\":{\"metrics\":null},\"1\":{\"metrics\":{\"metric\":[2,100,45,55,5]}}}}", app)
		assert.Equal(t, "{\"name\":\"default/test\",\"tracks\":{\"0\":\"0\",\"1\":\"1\"},\"versions\":{\"0\":{\"metrics\":{\"\":[0,0,0,0,0]}},\"1\":{\"metrics\":{\"metric\":[2,100,45,55,5]}}}}", app)
	}
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
