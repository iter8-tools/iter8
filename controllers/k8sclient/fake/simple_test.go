package fake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNew(t *testing.T) {
	var tests = []struct {
		a []runtime.Object
		b bool
	}{
		{nil, true},
		{[]runtime.Object{
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "iter8.tools",
					"kind":       "v1",
					"metadata": map[string]interface{}{
						"namespace": "myns",
						"name":      "myname",
					},
				},
			},
		}, true},
	}

	for _, e := range tests {
		client := New(nil, e.a)
		assert.NotNil(t, client)
	}
}
