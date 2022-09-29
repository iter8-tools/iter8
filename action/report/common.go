package report

import (
	"fmt"
	"sort"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

// Reporter implements methods that are common to text and HTML reporting.
type Reporter struct {
	// Experiment enables access to all base.Experiment data and methods
	*base.Experiment
}

// SortedScalarAndSLOMetrics extracts and sorts metric names from experiment.
// It looks for scalar metrics referenced in the MetricsInfo section,
// and also for scalar metrics referenced in SLOs.
func (r *Reporter) SortedScalarAndSLOMetrics() []string {
	keys := []string{}
	// add scalar and summary metrics referenced in MetricsInfo
	for k, mm := range r.Result.Insights.MetricsInfo {
		if mm.Type == base.CounterMetricType || mm.Type == base.GaugeMetricType {
			keys = append(keys, k)
		}
		if mm.Type == base.SummaryMetricType {
			for _, agg := range []base.AggregationType{
				base.CountAggregator,
				base.MeanAggregator,
				base.StdDevAggregator,
				base.MinAggregator,
				base.MaxAggregator} {
				keys = append(keys, k+"/"+string(agg))
			}
		}
	}
	// also add metrics referenced in SLOs
	// only scalar metrics can feature in SLOs (for now)
	if r.Result.Insights.SLOs != nil {
		for _, v := range r.Result.Insights.SLOs.Upper {
			nm, err := base.NormalizeMetricName(v.Metric)
			if err == nil {
				keys = append(keys, nm)
			}
		}
		for _, v := range r.Result.Insights.SLOs.Lower {
			nm, err := base.NormalizeMetricName(v.Metric)
			if err == nil {
				keys = append(keys, nm)
			}
		}
	}
	// remove duplicates
	tmp := base.Uniq(keys)
	uniqKeys := []string{}
	for _, val := range tmp {
		uniqKeys = append(uniqKeys, val.(string))
	}
	// return sorted metrics
	sort.Strings(uniqKeys)
	return uniqKeys
}

// ScalarMetricValueStr extracts value of a scalar metric (mn) for the given app version (j)
// Value is converted to string so that it can be printed in text and HTML reports.
func (r *Reporter) ScalarMetricValueStr(j int, mn string) string {
	val := r.Result.Insights.ScalarMetricValue(j, mn)
	if val != nil {
		return fmt.Sprintf("%0.2f", *val)
	}
	return "unavailable"
}

// MetricWithUnits provides the string representation of a metric name with units
func (r *Reporter) MetricWithUnits(metricName string) (string, error) {
	in := r.Result.Insights
	nm, err := base.NormalizeMetricName(metricName)
	if err != nil {
		return "", err
	}

	m, err := in.GetMetricsInfo(nm)
	if err != nil {
		e := fmt.Errorf("unable to get metrics info for %v", nm)
		log.Logger.Error(e)
		return "", e
	}
	str := nm
	if m.Units != nil {
		str = fmt.Sprintf("%v (%v)", str, *m.Units)
	}
	return str, nil
}
