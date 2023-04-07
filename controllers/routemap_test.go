package controllers

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// normalizeWeights sets the normalized weights for each variant of the routemap
// the inputs for normalizedWeights include:
// 1. Whether or not variants are available
// 2. Weights set using annotations
// 3. Weights set in the routemap for each variant
// 4. Default weight of 1 for each variant
func TestNormalizeWeights_sum_zero(t *testing.T) {
	// 1. Create a routemap with variants
	testRoutemap := routemap{
		mutex:      sync.RWMutex{},
		ObjectMeta: metav1.ObjectMeta{Name: "testRoutemap", Namespace: "default"},
		Variants: []variant{
			{
				Resources: []resource{{
					GVRShort:  "svc",
					Name:      "resource1",
					Namespace: base.StringPointer("default"),
				}},
			},
			{
				Resources: []resource{{
					GVRShort:  "svc",
					Name:      "resource2",
					Namespace: base.StringPointer("default"),
				}},
			},
			{
				Resources: []resource{{
					GVRShort:  "svc",
					Name:      "resource3",
					Namespace: base.StringPointer("default"),
				}},
			},
		},
	}

	// 2. Create config
	testConfig := &Config{
		ResourceTypes: map[string]GroupVersionResourceConditions{
			"svc": {
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "services",
				},
			},
		},
		DefaultResync: "30s",
	}

	// 3. For each entry in table driven tests
	// //	1. Create mock appinformers with variants in different states
	// // 2. Get normalize weight
	var tests = []struct {
		b []uint32
	}{
		{[]uint32{defaultVariantWeight, 0, 0}},
	}

	for _, e := range tests {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // cancel when we are finished consuming integers

		_ = initAppResourceInformers(ctx.Done(), testConfig, fake.New(nil, nil))
		testRoutemap.normalizeWeights(testConfig)
		assert.Equal(t, e.b, testRoutemap.Weights())
	}
}

func TestExtractRouteMap(t *testing.T) {
	// get config
	_ = os.Setenv(ConfigEnv, base.CompletePath("../", "testdata/controllers/config.yaml"))
	conf, err := readConfig()
	assert.NoError(t, err)

	// make cm
	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				iter8ManagedByLabel: iter8ManagedByValue,
				iter8KindLabel:      iter8KindRoutemapValue,
				iter8VersionLabel:   iter8VersionValue,
			},
		},
		Immutable: base.BoolPointer(true),
		Data: map[string]string{
			"strSpec": `
variants:
- resources: []
  weight: 1
`,
		},
		BinaryData: map[string][]byte{},
	}

	// get routemap from cm
	rm, err := extractRoutemap(&cm, conf)
	assert.NoError(t, err)
	assert.NotNil(t, rm)
}

func TestConditionsSatisfied(t *testing.T) {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	u.SetGeneration(13)
	config := &Config{
		ResourceTypes: map[string]GroupVersionResourceConditions{"foo": {
			Conditions: []Condition{{
				Name:   "bar",
				Status: "True",
			}},
		}},
	}
	var gen12 = int64(12)
	var gen13 = int64(13)

	var tests = []struct {
		conditions []interface{}
		satisfied  bool
	}{
		{nil, false},
		{[]interface{}{"a", "b"}, false},
		{[]interface{}{map[string]interface{}{
			"status": "got it",
		}}, false},
		{[]interface{}{map[string]interface{}{
			"type": "bar",
		}}, false},
		{[]interface{}{map[string]interface{}{
			"type":               "bar",
			"status":             "True",
			"observedGeneration": gen12,
		}}, false},
		{[]interface{}{map[string]interface{}{
			"type":               "bar",
			"status":             "False",
			"observedGeneration": gen13,
		}}, false},
		{[]interface{}{map[string]interface{}{
			"type":               "bar",
			"status":             "True",
			"observedGeneration": gen13,
		}}, true},
	}

	for _, tt := range tests {
		_ = unstructured.SetNestedMap(u.Object, make(map[string]interface{}), "status")
		_ = unstructured.SetNestedSlice(u.Object, tt.conditions, "status", "conditions")
		sat := conditionsSatisfied(u, "foo", config)
		assert.Equal(t, tt.satisfied, sat)
	}
}
