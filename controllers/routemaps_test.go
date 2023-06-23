package controllers

import (
	"sync"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRouteMaps_Delete(t *testing.T) {
	s := routemaps{
		mutex: sync.RWMutex{},
		nsRoutemap: map[string]routemapsByName{
			"default": {
				"test": {
					mutex:             sync.RWMutex{},
					ObjectMeta:        metav1.ObjectMeta{},
					Versions:          []version{},
					RoutingTemplates:  map[string]routingTemplate{},
					normalizedWeights: []uint32{},
				},
			},
		},
	}
	s.delete(&corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Immutable: base.BoolPointer(true),
	})
	obj, ok := s.nsRoutemap["default"]
	assert.False(t, ok)
	assert.Nil(t, obj)
}
