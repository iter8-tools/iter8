package application

import (
	"fmt"
	"math"
	"time"

	"github.com/iter8-tools/iter8/base/log"
)

type SummaryMetric [6]float64

const (
	COUNT_IDX = 0
	SUM_IDX   = 1
	MIN_IDX   = 2
	MAX_IDX   = 3
	SS_IDX    = 4
	TS_IDX    = 5
)

func EmptySummaryMetric() SummaryMetric {
	return SummaryMetric{
		0,                           // Count
		0,                           // Sum
		math.MaxFloat64,             // Min
		math.SmallestNonzeroFloat64, // Max
		0,                           // SumSquares
		float64(time.Now().Unix()),  // LastUpdateTimestamp
	}
}

func (m *SummaryMetric) Count() uint32 {
	return uint32(math.Round((*m)[COUNT_IDX]))
}

func (m *SummaryMetric) SetCount(v uint32) {
	(*m)[COUNT_IDX] = float64(v)
}

func (m *SummaryMetric) Sum() float64 {
	return (*m)[SUM_IDX]
}

func (m *SummaryMetric) SetSum(v float64) {
	(*m)[SUM_IDX] = v
}

func (m *SummaryMetric) Min() float64 {
	return (*m)[MIN_IDX]
}

func (m *SummaryMetric) SetMin(v float64) {
	if v < m.Min() {
		(*m)[MIN_IDX] = v
	}
}

func (m *SummaryMetric) Max() float64 {
	return (*m)[MAX_IDX]
}

func (m *SummaryMetric) SetMax(v float64) {
	if v > m.Max() {
		(*m)[MAX_IDX] = v
	}
}

func (m *SummaryMetric) SumSquares() float64 {
	return (*m)[SS_IDX]
}

func (m *SummaryMetric) SetSumSquares(v float64) {
	(*m)[SS_IDX] = v
}

func (m *SummaryMetric) LastUpdateTimestamp() time.Time {
	return time.Unix(int64(math.Round((*m)[TS_IDX])), 0)
}

func (m *SummaryMetric) SetLastUpdateTimestamp(t time.Time) {
	(*m)[TS_IDX] = float64(t.Unix())
}

func (metric SummaryMetric) Add(value float64) SummaryMetric {
	log.Logger.Tracef("Add() called with: <%s>", metric.toString())
	metric.SetCount(metric.Count() + 1)
	metric.SetSum(metric.Sum() + value)
	metric.SetMin(value)
	metric.SetMax(value)
	metric.SetSumSquares(metric.SumSquares() + (value * value))
	metric.SetLastUpdateTimestamp(time.Now())
	log.Logger.Tracef("after Add(): <%s>", metric.toString())
	return metric
}

func (metric *SummaryMetric) toString() string {
	return fmt.Sprintf("[%d] %f < %f, (%f, %f) %s",
		metric.Count(),
		metric.Min(),
		metric.Max(),
		metric.Sum(),
		metric.SumSquares(),
		metric.LastUpdateTimestamp(),
	)
}
