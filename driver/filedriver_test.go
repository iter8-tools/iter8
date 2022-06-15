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
	err := base.RunExperiment(false, &fd)
	assert.NoError(t, err)
	exp, err := base.BuildExperiment(&fd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure() && exp.SLOs())
}

func TestFileDriverReadError(t *testing.T) {
	base.SetupWithMock(t)

	fd := FileDriver{
		RunDir: ".",
	}

	exp, err := fd.Read()
	assert.Error(t, err)
	assert.Nil(t, exp)
}
