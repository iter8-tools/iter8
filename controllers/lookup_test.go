package controllers

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLookupInternal(t *testing.T) {

	allRoutemaps = routemaps{
		mutex: sync.RWMutex{},
		nsRoutemap: map[string]routemapsByName{
			"default": {
				"test": {
					mutex:             sync.RWMutex{},
					ObjectMeta:        metav1.ObjectMeta{},
					Versions:          make([]version, 2),
					RoutingTemplates:  map[string]routingTemplate{},
					normalizedWeights: []uint32{},
				},
			},
		},
	}

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
