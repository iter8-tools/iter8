package application

// summary_metric.go - defines a summary metric object. For space efficiency it is an array of 6 float64.

import (
	"fmt"
	"math"
	"time"
)

// SummaryMetric is summary metric type
type SummaryMetric [6]float64

const (
	COUNT_IDX = 0
	SUM_IDX   = 1
	MIN_IDX   = 2
	MAX_IDX   = 3
	SS_IDX    = 4
	TS_IDX    = 5
)

// EmptySummaryMetric returns metric object without any values added
func EmptySummaryMetric() *SummaryMetric {
	m := SummaryMetric{
		0,                           // Count
		0,                           // Sum
		math.MaxFloat64,             // Min
		math.SmallestNonzeroFloat64, // Max
		0,                           // SumSquares
		float64(time.Now().Unix()),  // LastUpdateTimestamp
	}
	return &m
}

// Count returns the number of observed values summarized by the metric
func (m *SummaryMetric) Count() uint32 {
	return uint32(math.Round((*m)[COUNT_IDX]))
}

// SetCount sets the number of observed values summarized by the metric
func (m *SummaryMetric) SetCount(v uint32) {
	(*m)[COUNT_IDX] = float64(v)
}

// Sum is the sum of the observed values
func (m *SummaryMetric) Sum() float64 {
	return (*m)[SUM_IDX]
}

// SetSum sets the sum of the observed values
func (m *SummaryMetric) SetSum(v float64) {
	(*m)[SUM_IDX] = v
}

// Min is the minimum of the observed values
func (m *SummaryMetric) Min() float64 {
	return (*m)[MIN_IDX]
}

// SetMin sets the minimum of the observed values
func (m *SummaryMetric) SetMin(v float64) {
	if v < m.Min() {
		(*m)[MIN_IDX] = v
	}
}

// Max is the maximum of the observed values
func (m *SummaryMetric) Max() float64 {
	return (*m)[MAX_IDX]
}

// SetMax sets the maximum of the observed values
func (m *SummaryMetric) SetMax(v float64) {
	if v > m.Max() {
		(*m)[MAX_IDX] = v
	}
}

// SumSquares is the sum of the squares of the observed values
func (m *SummaryMetric) SumSquares() float64 {
	return (*m)[SS_IDX]
}

// SetSumSquares sets the sum of the squares of the observed values
func (m *SummaryMetric) SetSumSquares(v float64) {
	(*m)[SS_IDX] = v
}

// LastUpdateTimestamp is the timestamp of the last change to the metric
func (m *SummaryMetric) LastUpdateTimestamp() time.Time {
	return time.Unix(int64(math.Round((*m)[TS_IDX])), 0)
}

// SetLastUpdateTimestamp sets the timestamp
func (m *SummaryMetric) SetLastUpdateTimestamp(t time.Time) {
	(*m)[TS_IDX] = float64(t.Unix())
}

// Add adds an observed value to the metric
func (m *SummaryMetric) Add(value float64) *SummaryMetric {
	m.SetCount(m.Count() + 1)
	m.SetSum(m.Sum() + value)
	m.SetMin(value)
	m.SetMax(value)
	m.SetSumSquares(m.SumSquares() + (value * value))
	m.SetLastUpdateTimestamp(time.Now())
	return m
}

// String returns a string representing the metric (not all fields are included)
func (m *SummaryMetric) String() string {
	return fmt.Sprintf("[%d] %f, %f, %f, %f",
		m.Count(),
		m.Min(),
		m.Max(),
		m.Sum(),
		m.SumSquares(),
	)
}
