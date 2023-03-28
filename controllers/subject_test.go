package controllers

import (
	"context"
	"sync"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// normalizeWeights sets the normalized weights for each variant of the subject
// the inputs for normalizedWeights include:
// 1. Whether or not variants are available
// 2. Weights set using annotations
// 3. Weights set in the subject for each variant
// 4. Default weight of 1 for each variant
func TestNormalizeWeights_sum_zero(t *testing.T) {
	// 1. Create a subject with variants
	testSubject := subject{
		mutex:      sync.RWMutex{},
		ObjectMeta: metav1.ObjectMeta{Name: "testSubject", Namespace: "default"},
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
				Group:    "",
				Version:  "v1",
				Resource: "services",
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

		initAppResourceInformers(ctx.Done(), testConfig, fake.New())
		testSubject.normalizeWeights(testConfig)
		assert.Equal(t, e.b, testSubject.Weights())
	}
}
