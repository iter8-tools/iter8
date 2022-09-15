package application

// summary_metric.go - defines a summary metric object. For space efficiency it is an array of 6 float64.

import (
	"fmt"
	"math"
)

// SummaryMetric is summary metric type
type SummaryMetric [5]float64

const (
	countIdx = 0
	sumIdx   = 1
	minIdx   = 2
	maxIdx   = 3
	ssIdx    = 4
)

// EmptySummaryMetric returns metric object without any values added
func EmptySummaryMetric() *SummaryMetric {
	m := SummaryMetric{
		0,                           // Count
		0,                           // Sum
		math.MaxFloat64,             // Min
		math.SmallestNonzeroFloat64, // Max
		0,                           // SumSquares
	}
	return &m
}

// Count returns the number of observed values summarized by the metric
func (m *SummaryMetric) Count() uint32 {
	return uint32(math.Round((*m)[countIdx]))
}

// SetCount sets the number of observed values summarized by the metric
func (m *SummaryMetric) SetCount(v uint32) {
	(*m)[countIdx] = float64(v)
}

// Sum is the sum of the observed values
func (m *SummaryMetric) Sum() float64 {
	return (*m)[sumIdx]
}

// SetSum sets the sum of the observed values
func (m *SummaryMetric) SetSum(v float64) {
	(*m)[sumIdx] = v
}

// Min is the minimum of the observed values
func (m *SummaryMetric) Min() float64 {
	return (*m)[minIdx]
}

// SetMin sets the minimum of the observed values
func (m *SummaryMetric) SetMin(v float64) {
	if v < m.Min() {
		(*m)[minIdx] = v
	}
}

// Max is the maximum of the observed values
func (m *SummaryMetric) Max() float64 {
	return (*m)[maxIdx]
}

// SetMax sets the maximum of the observed values
func (m *SummaryMetric) SetMax(v float64) {
	if v > m.Max() {
		(*m)[maxIdx] = v
	}
}

// SumSquares is the sum of the squares of the observed values
func (m *SummaryMetric) SumSquares() float64 {
	return (*m)[ssIdx]
}

// SetSumSquares sets the sum of the squares of the observed values
func (m *SummaryMetric) SetSumSquares(v float64) {
	(*m)[ssIdx] = v
}

// Add adds an observed value to the metric
func (m *SummaryMetric) Add(value float64) *SummaryMetric {
	m.SetCount(m.Count() + 1)
	m.SetSum(m.Sum() + value)
	m.SetMin(value)
	m.SetMax(value)
	m.SetSumSquares(m.SumSquares() + (value * value))
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
