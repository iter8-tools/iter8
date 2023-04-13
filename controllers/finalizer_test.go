package controllers

import (
	"context"
	"fmt"
	"testing"

	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAddFinalizer(t *testing.T) {

	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"namespace": "myns",
				"name":      "myname",
			},
		},
	}

	// 1. Create a test k8sclient.Interface object with a k8s resource
	client := fake.New(nil, []runtime.Object{u})

	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	// 2. Config must have the gvrshort for that resource
	config := &Config{
		ResourceTypes: map[string]GroupVersionResourceConditions{
			"pod": {
				GroupVersionResource: gvr,
			}},
	}

	// 3. Call add finalizer
	addFinalizer("myname", "myns", "pod", client, config)

	// 4. Get obj from client
	u1, err := client.FakeDynamicClient.Resource(gvr).Namespace("myns").Get(context.Background(), "myname", v1.GetOptions{})

	// 5. Test for finalizer string
	assert.NoError(t, err)
	assert.Contains(t, u1.GetFinalizers(), iter8FinalizerStr)
}

func TestRemoveFinalizer(t *testing.T) {

	type test struct {
		u                     *unstructured.Unstructured
		contains              bool
		expectedNumFinalizers int
	}

	tt := []test{
		{&unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"namespace":  "myns",
					"name":       "myname1",
					"finalizers": []any{iter8FinalizerStr},
				},
			},
		}, true, 1}, // deletionTimestamp not present --> no deletion
		{&unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"namespace":         "myns",
					"name":              "myname1",
					"finalizers":        []any{"notIter8Finalizer"},
					"deletionTimestamp": "2020-10-22T21:30:34Z",
				},
			},
		}, false, 1}, // not Iter8 finalizer --> no deletion
		{&unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"namespace":         "myns",
					"name":              "myname2",
					"finalizers":        []any{iter8FinalizerStr},
					"deletionTimestamp": "2020-10-22T21:30:34Z",
				},
			},
		}, false, 0}} // Iter8 finalizer + deletionTimestamp --> deletion

	// pod gvr
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	// Config must have the gvrshort for that resource
	config := &Config{
		ResourceTypes: map[string]GroupVersionResourceConditions{
			"pod": {
				GroupVersionResource: gvr,
			}},
	}

	for i, tc := range tt {
		// 1. Create a test k8sclient.Interface object with a k8s resource
		client := fake.New(nil, []runtime.Object{tc.u})

		// 2. Call remove finalizer
		removeFinalizer(tc.u.GetName(), tc.u.GetNamespace(), "pod", client, config)

		// 3. Get obj from client
		u1, err := client.FakeDynamicClient.Resource(gvr).Namespace(tc.u.GetNamespace()).Get(context.Background(), tc.u.GetName(), v1.GetOptions{})

		// 4. Test for finalizer string
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedNumFinalizers, len(u1.GetFinalizers()), fmt.Sprintf("iteration: %d", i))
		if tc.contains {
			assert.Contains(t, u1.GetFinalizers(), iter8FinalizerStr, fmt.Sprintf("iteration: %d", i))
		} else {
			assert.NotContains(t, u1.GetFinalizers(), iter8FinalizerStr, fmt.Sprintf("iteration: %d", i))
		}
	}
}
