package driver

import (
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestExperimentFromBytes(t *testing.T) {
	experiment := base.Experiment{}
	experimentBytes, err := yaml.Marshal(experiment)
	assert.NoError(t, err)
	assert.NotNil(t, experimentBytes)

	// Experiment from marshalled experiment
	experiment2, err := ExperimentFromBytes(experimentBytes)
	assert.NoError(t, err)
	assert.NotNil(t, experiment2)

	// Experiment from random bytes
	experiment3, err := ExperimentFromBytes([]byte{1, 2, 3})
	assert.Error(t, err)
	assert.Nil(t, experiment3)
}
