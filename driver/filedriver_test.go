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
	err := base.RunExperiment(&fd)
	assert.NoError(t, err)
	exp, err := base.BuildExperiment(true, &fd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure() && exp.SLOs())
}
