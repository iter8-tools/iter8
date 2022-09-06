package application

// version.go - supports notion of version of an application

import (
	"fmt"
	"strings"

	metrics "github.com/iter8-tools/iter8/base/metrics"
)

// Version is information about versions of an application in a Kubernetes cluster
type Version struct {
	// Track is track label assigned to this version
	// indicated by value of annotation iter8.tools/track
	Track *string `json:"track,omitempty" yaml:"metrics,omitempty"`
	// List of (summary) metrics for a version
	Metrics map[string]*metrics.SummaryMetric `json:"metrics" yaml:"metrics"`
}

// GetTrack returns a track identifier, if any. Otherwise, it returns nil.
func (v *Version) GetTrack() *string {
	return v.Track
}

func (v *Version) SetTrack(track *string) {
	v.Track = track
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

	track := "<no track>"
	if v.GetTrack() != nil {
		track = *v.GetTrack()
	}

	return fmt.Sprintf("\n\t%s\n\t%s",
		track,
		"- metrics: ["+strings.Join(metrics, ",")+"]",
	)
}
