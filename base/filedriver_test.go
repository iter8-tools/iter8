package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalRun(t *testing.T) {
	SetupWithMock(t)

	fd := FileDriver{
		RunDir: CompletePath("../", "testdata/drivertests"),
	}
	err := RunExperiment(&fd)
	assert.NoError(t, err)
	exp, err := BuildExperiment(true, &fd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure() && exp.SLOs())
}
