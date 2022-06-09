package driver

import (
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
)

func TestLocalRun(t *testing.T) {
	base.SetupWithMock(t)

	fd := FileDriver{
		RunDir: base.CompletePath("../", "testdata/drivertests"),
	}
	err := base.RunExperiment(&fd, false)
	assert.NoError(t, err)
	exp, err := base.BuildExperiment(true, &fd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure() && exp.SLOs())
}

func TestFileDriverReadMetricsSpec(t *testing.T) {
	base.SetupWithMock(t)

	fd := FileDriver{
		RunDir: base.CompletePath("../", "testdata/metrics"),
	}

	metrics, err := fd.ReadMetricsSpec("test-ce")

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
}

func TestFileDriverReadError(t *testing.T) {
	base.SetupWithMock(t)

	fd := FileDriver{
		RunDir: ".",
	}

	spec, err := fd.ReadSpec()
	assert.Error(t, err)
	assert.Nil(t, spec)

	metrics, err := fd.ReadMetricsSpec("test-ce")
	assert.Error(t, err)
	assert.Nil(t, metrics)

	result, err := fd.ReadResult()
	assert.Error(t, err)
	assert.Nil(t, result)
}
