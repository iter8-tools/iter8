package controllers

import (
	"sync"
	"testing"

	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestLookupInternal(t *testing.T) {
	namespace, name := "default", "test"

	// setup: add desired routemaps to allRoutemaps
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

func TestWriteMetricInternal(t *testing.T) {
	namespace, name := "default", "test"
	// setup: add desired routemaps to allRoutemaps
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

	// setup: create secret
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}

	// setup: create a (fake) client
	client := fake.New([]runtime.Object{&secret}, nil)

	// when we create a secret, Data[secretKey] is nil
	s, err := client.GetSecret(namespace, name)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	rawData := s.Data[secretKey]
	assert.Nil(t, rawData)

	err = writeMetricInternal(namespace+"/"+name, "user", "metric", "45", client)
	assert.NoError(t, err)

	// as a side effect, one of the versions of the routemap will have a metric value
	rm := allRoutemaps.getRoutemapFromNamespaceName(namespace, name)
	assert.NotNil(t, rm)
	assert.Equal(t, 1, len(rm.Versions[0].Metrics)+len(rm.Versions[1].Metrics))

	// verify we have written to the secret: Data[secretKey] is no longer empty
	s, err = client.GetSecret(namespace, name)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	rawData = s.Data[secretKey]
	assert.NotNil(t, rawData)

	// add a second value to metric
	err = writeMetricInternal(namespace+"/"+name, "user", "metric", "55", client)
	assert.NoError(t, err)

	// verify that we updated it correctly
	rm = allRoutemaps.getRoutemapFromNamespaceName(namespace, name)
	assert.NotNil(t, rm)
	assert.Equal(t, 1, len(rm.Versions[0].Metrics)+len(rm.Versions[1].Metrics))
	for i := range [2]int{} {
		if rm.Versions[i].Metrics != nil {
			m := rm.Versions[0].Metrics["metric"]
			assert.Equal(t, 2, m.Count())
			assert.Equal(t, 45, m.Min())
			assert.Equal(t, 55, m.Max())
			assert.Equal(t, 100, m.Sum())
		}
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
