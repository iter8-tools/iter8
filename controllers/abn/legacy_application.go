package abn

import (
	log "github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/base/summarymetrics"
)

// LegacyApplication is an legacy object capturing application details
// Deprecated: LegacyApplication will be removed when support for alternative metric stores is added.
type LegacyApplication struct {
	// Name is of the form namespace/Name where the Name is the value of the label app.kubernetes.io/Name
	Name string `json:"name" yaml:"name"`
	// Tracks is map from application track identifier to version name
	Tracks LegacyTracks `json:"tracks" yaml:"tracks"`
	// Versions maps version name to version data (a set of metrics)
	Versions LegacyVersions `json:"versions" yaml:"versions"`
}

// LegacyVersions is a map of the version name to a version object
// Deprecated: LegacyVersions will be removed when support for alternative metric stores is added.
type LegacyVersions map[string]*LegacyVersion

// LegacyTracks is map of track identifiers to version names
// Deprecated: LegacyTracks will be removed when support for alternative metric stores is added.
type LegacyTracks map[string]string

// LegacyVersion is information about versions of an application in a Kubernetes cluster
// Deprecated: LegacyVersion will be removed when support for alternative metric stores is added.
type LegacyVersion struct {
	// List of (summary) metrics for a version
	Metrics map[string]*summarymetrics.SummaryMetric `json:"metrics" yaml:"metrics"`
}

// GetVersion returns the Version object corresponding to a given version name
// If no corresponding version object exists, a new one will be created when allowNew is set to true
// returns the version object and a boolean indicating whether or not a new version was created or not
func (a *LegacyApplication) GetVersion(version string, allowNew bool) (*LegacyVersion, bool) {
	v, ok := a.Versions[version]
	if !ok {
		if allowNew {
			v = &LegacyVersion{
				Metrics: map[string]*summarymetrics.SummaryMetric{},
			}
			log.Logger.Debugf("GetVersion no data found; creating %+v", v)
			a.Versions[version] = v
			return v, true
		}
		return nil, false
	}

	log.Logger.Debugf("GetVersion returning %+v", v)
	return v, false
}
