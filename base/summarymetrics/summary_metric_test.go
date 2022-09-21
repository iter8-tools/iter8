package summarymetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionAndSummaryMetric(t *testing.T) {
	m := EmptySummaryMetric()

	assert.NotNil(t, m)
	// new metric is empty
	assert.Equal(t, uint32(0), m.Count())
	assert.Equal(t, float64(0), m.Sum())

	// add values
	m.Add(float64(27))
	m.Add(float64(56))
	assert.Equal(t, uint32(2), m.Count())
	assert.Equal(t, float64(27), m.Min())
	assert.Equal(t, float64(56), m.Max())
	assert.Equal(t, float64(83), m.Sum())
	assert.Equal(t, float64(3865), m.SumSquares())
	assert.Equal(t, "[2] 27.000000, 56.000000, 83.000000, 3865.000000", m.String())
}
