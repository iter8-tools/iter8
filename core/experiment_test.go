package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecRead(t *testing.T) {

	filePath = CompletePath("../", "testdata/spec.yaml")

	e, err := Read()
	Logger.Info(e)
	assert.NoError(t, err)
	assert.Equal(t, "collect-fortio-metrics", *e.Spec.Tasks[0].Task)
}
