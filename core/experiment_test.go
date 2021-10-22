package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecRead(t *testing.T) {

	fc := FileContext{
		SpecFile:   CompletePath("../", "testdata/spec.yaml"),
		ResultFile: "",
	}

	es, err := fc.ReadSpec()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(es.Tasks))
	assert.Equal(t, "collect-fortio-metrics", es.Tasks[0].Task)
}
