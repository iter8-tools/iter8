package controllers

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// tests that we get the same result for the same inputs
func TestLookupInternal(t *testing.T) {
	namespace, name := "default", "test"

	// setup: add desired routemaps to allRoutemaps; first clear all routemaps
	allRoutemaps = routemaps{
		nsRoutemap: make(map[string]routemapsByName),
	}
	addRouteMapForTest(namespace, name, &routemap{
		mutex: sync.RWMutex{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Versions:          make([]version, 2),
		RoutingTemplates:  map[string]routingTemplate{},
		normalizedWeights: []uint32{},
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

func addRouteMapForTest(ns string, n string, s *routemap) {
	// make sure nsRoutemap has entry for namespace ns
	_, ok := allRoutemaps.nsRoutemap[ns]
	if !ok {
		allRoutemaps.nsRoutemap[ns] = make(map[string]*routemap)
	}
	// add (or replace) routemap for n
	(allRoutemaps.nsRoutemap[ns])[n] = s
}
