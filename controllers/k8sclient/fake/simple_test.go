package fake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	myns    = "myns"
	myname  = "myname"
	myname2 = "myname2"
	hello   = "hello"
	world   = "world"
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
						"namespace": myns,
						"name":      myname,
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

func TestPatch(t *testing.T) {
	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	client := New(nil, []runtime.Object{
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"namespace": myns,
					"name":      myname,
				},
			},
		},
	})

	myDeployment, err := client.FakeDynamicClient.Resource(gvr).Namespace(myns).Get(context.TODO(), myname, v1.GetOptions{})

	assert.NoError(t, err)
	assert.NotNil(t, myDeployment)
	// myDeployment should not have the hello: world label yet
	assert.Equal(t, "", myDeployment.GetLabels()[hello])

	// Create a copy of myDeployment and add the hello: world label
	copiedDeployment := myDeployment.DeepCopy()
	copiedDeployment.SetLabels(map[string]string{
		hello: world,
	})
	newDeploymentBytes, err := copiedDeployment.MarshalJSON()
	assert.NoError(t, err)
	assert.NotNil(t, newDeploymentBytes)

	// Patch myDeployment
	patchedDeployment, err := client.Patch(gvr, myns, myname, newDeploymentBytes)

	assert.NoError(t, err)
	assert.NotNil(t, patchedDeployment)
	// Patched myDeployment should now have the hello: world label
	assert.Equal(t, world, patchedDeployment.GetLabels()[hello])
}
