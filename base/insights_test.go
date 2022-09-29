package base

import (
	"testing"

	"github.com/iter8-tools/iter8/base/summarymetrics"
	"github.com/stretchr/testify/assert"
)

func TestTrackVersionStr(t *testing.T) {
	scenarios := map[string]struct {
		in          Insights
		expectedStr string
	}{
		"VersionNames is nil":         {in: Insights{}, expectedStr: "version 0"},
		"Version and Track empty":     {in: Insights{VersionNames: []VersionInfo{}}, expectedStr: "version 0"},
		"Track is empty":              {in: Insights{VersionNames: []VersionInfo{{Version: "version"}}}, expectedStr: "version"},
		"Version is empty":            {in: Insights{VersionNames: []VersionInfo{{Track: "track"}}}, expectedStr: "track"},
		"Version and Track not empty": {in: Insights{VersionNames: []VersionInfo{{Track: "track", Version: "version"}}}, expectedStr: "track (version)"},
	}

	for l, s := range scenarios {
		t.Run(l, func(t *testing.T) {
			assert.Equal(t, s.expectedStr, s.in.TrackVersionStr(0))
		})
	}
}

func TestGetSummaryAggregation(t *testing.T) {
	in := Insights{
		// count, sum, min, max, sumsquares
		SummaryMetricValues: []map[string]summarymetrics.SummaryMetric{{
			"metric": [5]float64{float64(10), float64(110), float64(2), float64(20), float64(1540)},
		}},
	}

	assert.Equal(t, float64(10), *in.getSummaryAggregation(0, "metric", "count"))
	assert.Equal(t, float64(11), *in.getSummaryAggregation(0, "metric", "mean"))
	// assert.Equal(t, float64(6.055300708194983), *in.getSummaryAggregation(0, "metric", "stddev"))
	assert.Greater(t, float64(6.0553008), *in.getSummaryAggregation(0, "metric", "stddev"))
	assert.Less(t, float64(6.0553007), *in.getSummaryAggregation(0, "metric", "stddev"))
	assert.Equal(t, float64(2), *in.getSummaryAggregation(0, "metric", "min"))
	assert.Equal(t, float64(20), *in.getSummaryAggregation(0, "metric", "max"))

	assert.Nil(t, in.getSummaryAggregation(0, "metric", "invalid"))

	assert.Nil(t, in.getSummaryAggregation(0, "notametric", "count"))
}

func TestGetSampleAggregation(t *testing.T) {
	// no values
	in := Insights{
		NonHistMetricValues: []map[string][]float64{{
			"metric": []float64{},
		}},
	}
	assert.Nil(t, in.getSampleAggregation(0, "metric", "something"))

	// single value
	in = Insights{
		NonHistMetricValues: []map[string][]float64{{
			"metric": []float64{float64(2)},
		}},
	}
	assert.Equal(t, float64(2), *in.getSampleAggregation(0, "metric", "anything"))

	// multiple values
	in = Insights{
		NonHistMetricValues: []map[string][]float64{{
			"metric": []float64{
				float64(2), float64(4), float64(6), float64(8), float64(10),
				float64(12), float64(14), float64(16), float64(18), float64(20),
			},
		}},
	}
	assert.Len(t, in.NonHistMetricValues, 1)
	assert.Len(t, in.NonHistMetricValues[0], 1)
	assert.Contains(t, in.NonHistMetricValues[0], "metric")
	assert.Equal(t, float64(11), *in.getSampleAggregation(0, "metric", "mean"))
	// assert.Equal(t, float64(5.744562646538029), *in.getSampleAggregation(0, "metric", "stddev"))
	assert.Greater(t, float64(5.7445627), *in.getSampleAggregation(0, "metric", "stddev"))
	assert.Less(t, float64(5.7445626), *in.getSampleAggregation(0, "metric", "stddev"))
	assert.Equal(t, float64(2), *in.getSampleAggregation(0, "metric", "min"))
	assert.Equal(t, float64(20), *in.getSampleAggregation(0, "metric", "max"))
	// starts with p but not a percentile
	assert.Nil(t, in.getSampleAggregation(0, "metric", "p-notpercent"))
	// invalid percentile (101)
	assert.Nil(t, in.getSampleAggregation(0, "metric", "p101"))
	assert.Equal(t, float64(15), *in.getSampleAggregation(0, "metric", "p78.3"))
	// not a valid aggregation
	assert.Nil(t, in.getSampleAggregation(0, "metric", "invalid"))
}

func TestAggregateMetric(t *testing.T) {
	in := Insights{
		MetricsInfo: map[string]MetricMeta{
			"prefix/summary": {Type: SummaryMetricType},
			"prefix/sample":  {Type: SampleMetricType},
			"prefix/counter": {Type: CounterMetricType},
			"prefix/gauge":   {Type: GaugeMetricType},
		},
		NonHistMetricValues: []map[string][]float64{{
			"prefix/sample": []float64{
				float64(2), float64(4), float64(6), float64(8), float64(10),
				float64(12), float64(14), float64(16), float64(18), float64(20),
			},
		}},
		// count, sum, min, max, sumsquares
		SummaryMetricValues: []map[string]summarymetrics.SummaryMetric{{
			"prefix/summary": [5]float64{float64(10), float64(110), float64(2), float64(20), float64(1540)},
		}},
	}

	// not enough parts
	assert.Nil(t, in.aggregateMetric(0, "counter"))
	// not enough parts
	assert.Nil(t, in.aggregateMetric(0, "prefix/counter"))
	// not a summary or sample metric
	assert.Nil(t, in.aggregateMetric(0, "prefix/counter/mean"))
	// not in MetricsInfo
	assert.Nil(t, in.aggregateMetric(0, "prefix/invalid/mean"))

	assert.Equal(t, float64(11), *in.aggregateMetric(0, "prefix/summary/mean"))
	assert.Equal(t, float64(11), *in.aggregateMetric(0, "prefix/sample/mean"))
}
