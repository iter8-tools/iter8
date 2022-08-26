package application

// version.go - supports notion of version of an application

import (
	"fmt"
	"strings"
	"time"
)

// Version is information about versions of an application in a Kubernetes cluster
type Version struct {
	// Ready is indicator that the version is ready to receive traffic
	// indicated by value of annotation iter8.tools/ready
	Ready bool `json:"ready" yaml:"ready"`
	// Track is track label assigned to this version
	// indicated by value of annotation iter8.tools/track
	Track *string `json:"track,omitempty" yaml:"metrics,omitempty"`
	// List of (summary) metrics for a version
	Metrics map[string]*SummaryMetric `json:"metrics" yaml:"metrics"`
	// LastUpdateTimestamp is time of last update (either event or metric)
	LastUpdateTimestamp time.Time `json:"lastUpdateTimestamp" yaml:"lastUpdateTimestamp"`
}

// GetTrack returns a track identifier, if any. Otherwise, it returns nil.
func (v *Version) GetTrack() *string {
	return v.Track
}

// IsReady determines if the version is ready
// Inspects history in reverse to determine last relevant event
func (v *Version) IsReady() bool {
	return v.Ready
}

// GetMetric returns a metric from the list of metrics associated with a version
// If no metric is present for a given name, a new one is created
func (v *Version) GetMetric(metric string, allowNew bool) (*SummaryMetric, bool) {
	m, ok := v.Metrics[metric]
	if !ok {
		if allowNew {
			m := EmptySummaryMetric()
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

	return fmt.Sprintf("\n\t%t %s\n\t%s",
		v.IsReady(), track,
		"- metrics: ["+strings.Join(metrics, ",")+"]",
	)
}
