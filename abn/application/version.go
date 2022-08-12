package application

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base/log"
)

// Version is information about versions of an application in a Kubernetes cluster
type Version struct {
	// History is a time ordered list of events that have been observed
	History []VersionEvent `json:"history" yaml:"history"`
	// List of (summary) metrics for a version
	Metrics map[string]*SummaryMetric `json:"metrics" yaml:"metrics"`
	// LastUpdateTimestamp is time of last update (either event or metric)
	LastUpdateTimestamp time.Time `json:"lastUpdateTimestamp" yaml:"lastUpdateTimestamp"`
}

// VersionEvent is a record of an observed event of interest
type VersionEvent struct {
	// Timestamp is the time the event is observed
	Timestamp time.Time `json:"tm" yaml:"tm"`
	// Type is the type of the event
	Type VersionEventType `json:"ev" yaml:"ev"`
	// Track is a track identifier (parameter of a VersionMapTrackEvent)
	Track string `json:"trk,omitempty" yaml:"trk,omitempty"`
}

// VersionEventType is type of a VersionEvent type
type VersionEventType string

const (
	// VersionNewEvent indicates a new version has been identified
	VersionNewEvent VersionEventType = "new"
	// VersionReadyEvetn indicates that a version (that was not previously noted as ready) is ready
	VersionReadyEvent VersionEventType = "ready"
	// VersionNoLongerReadyEvent indicates that a version (that was previously noted as ready) is not ready
	VersionNoLongerReadyEvent VersionEventType = "notready"
	// VersionMapTrackEvent indicates that a version (that was not previously associated with a track identifier) is associated with a track identifier
	VersionMapTrackEvent VersionEventType = "track"
	// VersionUnmapTrackEvent indicates that version (that was previously noted as being associated with a track identifier) is not longer associated wtih a track identifier
	// This happens when a version is no longer ready or a track is assigned to a different version
	VersionUnmapTrackEvent VersionEventType = "untrack"
)

// GetTrack returns a track identifier, if any. Otherwise, it returns nil.
func (v *Version) GetTrack() *string {
	numEvents := len(v.History)
	if numEvents == 0 {
		return nil
	}
	lastEvent := v.History[numEvents-1]
	if lastEvent.Type == VersionMapTrackEvent {
		return &lastEvent.Track
	}
	return nil
}

// AddEvent adds an event to the version history
func (v *Version) AddEvent(typ VersionEventType, track ...string) error {
	log.Logger.Tracef("AddEvent() called with event %s", typ)

	e := VersionEvent{
		Type:      typ,
		Timestamp: time.Now(),
	}
	if typ == VersionMapTrackEvent {
		if len(track) != 1 {
			return errors.New("map track event requires track")
		}
		e.Track = track[0]
	}

	v.History = append(v.History, e)
	return nil
}

// IsReady determines if the version is ready
// Inspects history in reverse to determine last relevant event
func (v *Version) IsReady() bool {
	// numEvents = len(v.History)
	for i := len(v.History) - 1; i >= 0; i-- {
		switch v.History[i].Type {
		case VersionReadyEvent:
			return true
		case VersionNoLongerReadyEvent:
			return false
		}
	}
	return false
}

// GetMetric returns a metric from the list of metrics associated with a version
// If no metric is present for a given name, a new one is created
func (v *Version) GetMetric(metric string, allowNew bool) (*SummaryMetric, bool) {
	log.Logger.Tracef("GetMetric(%s) called", metric)
	log.Logger.Tracef("version (before) is %s", *v)
	m, ok := v.Metrics[metric]
	if !ok {
		if allowNew {
			newM := EmptySummaryMetric()
			v.Metrics[metric] = &newM
			log.Logger.Tracef("version (end) is %s", *v)
			return v.Metrics[metric], true
		} else {
			log.Logger.Tracef("version (end) is %s", *v)
			return nil, false
		}
	}
	log.Logger.Tracef("version (end) is %s", *v)
	return m, false
}

func (v *Version) String() string {
	types := []string{}
	for _, e := range v.History {
		types = append(types, string(e.Type))
	}

	metrics := []string{}
	for n, _ := range v.Metrics {
		metrics = append(metrics, n)
	}

	return fmt.Sprintf("\n\t%s\n\t%s",
		"- history: ["+strings.Join(types, ",")+"]",
		"- metrics: ["+strings.Join(metrics, ",")+"]",
	)
}
