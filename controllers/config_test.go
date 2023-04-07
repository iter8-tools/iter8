package controllers

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestReadConfig(t *testing.T) {
	var tests = []struct {
		confEnv  bool
		confFile string
		valid    bool
	}{
		{true, base.CompletePath("../", "testdata/controllers/config.yaml"), true},
		{false, base.CompletePath("../", "testdata/controllers/config.yaml"), false},
		{true, base.CompletePath("../", "testdata/controllers/garb.age"), false},
		{true, base.CompletePath("../", "this/file/does/not/exist"), false},
	}

	for _, tt := range tests {
		_ = os.Unsetenv(ConfigEnv)
		if tt.confEnv {
			_ = os.Setenv(ConfigEnv, tt.confFile)
		}

		c, err := readConfig()
		if tt.valid {
			assert.NoError(t, err)
			assert.Equal(t, "30s", c.DefaultResync)
			assert.Equal(t, 5, len(c.ResourceTypes))
			isvc := c.ResourceTypes["isvc"]
			assert.Equal(t, isvc, GroupVersionResourceConditions{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "serving.kserve.io",
					Version:  "v1beta1",
					Resource: "inferenceservices",
				},
				Conditions: []Condition{{
					Name:   "Ready",
					Status: "True",
				}},
			})
		} else {
			assert.Error(t, err)
		}
	}
}
