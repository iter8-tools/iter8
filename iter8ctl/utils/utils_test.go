package utils

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletePath(t *testing.T) {
	p1 := CompletePath("", "a")
	p2 := CompletePath("../", "utils/a")
	p3 := CompletePath("", "b")
	assert.Equal(t, p1, p2)
	assert.NotEqual(t, p2, p3)
}

func ExampleCompletePath() {
	// Tests for the experiment package use code similar to the following snippet.
	filePath := CompletePath("../testdata", "experiment2.yaml")
	_, _ = ioutil.ReadFile(filePath)
}
