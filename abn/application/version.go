package application

// version.go - supports notion of version of an application

import (
	"fmt"
	"strings"

	metrics "github.com/iter8-tools/iter8/base/metrics"
)

// Version is information about versions of an application in a Kubernetes cluster
type Version struct {
	// List of (summary) metrics for a version
	Metrics map[string]*metrics.SummaryMetric `json:"metrics" yaml:"metrics"`
}

// GetMetric returns a metric from the list of metrics associated with a version
// If no metric is present for a given name, a new one is created
func (v *Version) GetMetric(metric string, allowNew bool) (*metrics.SummaryMetric, bool) {
	m, ok := v.Metrics[metric]
	if !ok {
		if allowNew {
			m := metrics.EmptySummaryMetric()
			v.Metrics[metric] = m
			return m, true
		} else {
			return nil, false
		}
	}
	return m, false
}

func (v *Version) String() string {
	metrics := []string{}
	for n, m := range v.Metrics {
		metrics = append(metrics, fmt.Sprintf("%s(%d)", n, m.Count()))
	}

	return fmt.Sprintf("\n%s",
		"- metrics: ["+strings.Join(metrics, ",")+"]",
	)
}
