package report

import (
	"fmt"
	"sort"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

// Reporter is a type that supports methods used for text and html reporting.
type Reporter struct {
	*base.Experiment
}

/* Following functions/methods are common to both text and html templates */

// SortedScalarAndSLOMetrics extracts and sorts metric names from experiment.
// It looks for available metrics in the results, and also for metrics specified in SLOs.
func (r *Reporter) SortedScalarAndSLOMetrics() []string {
	keys := []string{}
	for k, mm := range r.Result.Insights.MetricsInfo {
		if mm.Type == base.CounterMetricType || mm.Type == base.GaugeMetricType {
			keys = append(keys, k)
		}
	}
	// also add SLO metric names
	for _, v := range r.Result.Insights.SLOs {
		nm, err := base.NormalizeMetricName(v.Metric)
		if err == nil {
			keys = append(keys, nm)
		}
	}
	// remove duplicates
	tmp := base.Uniq(keys)
	uniqKeys := []string{}
	for _, val := range tmp {
		uniqKeys = append(uniqKeys, val.(string))
	}

	sort.Strings(uniqKeys)
	return uniqKeys
}

// ScalarMetricValueStr extracts the metric value string for the given version and given scalar metric name
func (r *Reporter) ScalarMetricValueStr(j int, mn string) string {
	val := r.Result.Insights.ScalarMetricValue(j, mn)
	if val != nil {
		return fmt.Sprintf("%0.2f", *val)
	} else {
		return "unavailable"
	}
}

// MetricWithUnits provides the string representation of metric name and with units
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
