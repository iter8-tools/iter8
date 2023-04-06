package controllers

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestReadConfig(t *testing.T) {
	_ = os.Setenv(configEnv, base.CompletePath("../", "testdata/controllers/config.yaml"))
	c, e := readConfig()
	assert.NoError(t, e)
	assert.Equal(t, "30s", c.DefaultResync)
	assert.Equal(t, 4, len(c.ResourceTypes))
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
}
