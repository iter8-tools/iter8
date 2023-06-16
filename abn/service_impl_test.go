package abn

import (
	"testing"

	"github.com/iter8-tools/iter8/controllers"
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
		Versions: make([]controllers.Version, 2),
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
	client := setupRouteMaps(t, namespace, name)

	// add a metric value
	err := writeMetricInternal(namespace+"/"+name, "user", "metric", "45", client)
	assert.NoError(t, err)

	// add a second value to metric
	err = writeMetricInternal(namespace+"/"+name, "user", "metric", "55", client)
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
		assert.Equal(t, "{\"name\":\"default/test\",\"tracks\":{\"0\":\"0\",\"1\":\"1\"},\"versions\":{\"0\":{\"metrics\":{\"metric\":[2,100,45,55,5050]}},\"1\":{\"metrics\":null}}}", app)
	} else {
		assert.Equal(t, "{\"name\":\"default/test\",\"tracks\":{\"0\":\"0\",\"1\":\"1\"},\"versions\":{\"0\":{\"metrics\":null},\"1\":{\"metrics\":{\"metric\":[2,100,45,55,5050]}}}}", app)
	}
}
