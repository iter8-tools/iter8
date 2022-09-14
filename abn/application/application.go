package application

// application.go - type of application

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/iter8-tools/iter8/base/log"
)

// Application is an application observed in a kubernetes cluster
type Application struct {
	// Name is of the form namespace/name where the name is the value of the label app.kubernetes.io/name
	Name string `json:"name" yaml:"name"`
	// Tracks is map from application track identifier to version name
	Tracks `json:"tracks" yaml:"tracks"`
	// Versions is a map of versions name to version data
	Versions `json:"versions" yaml:"versions"`
}

// Versions is a map of the version name to a version object
type Versions map[string]*Version

// Tracks is map of track identifiers to version names
type Tracks map[string]string

// GetVersion returns the Version object corresponding to a given version name
// If no corresponding version object exists, a new one will be created when allowNew is set to true
// returns the version object and a boolean indicating whether or not a new version was created or not
func (a *Application) GetVersion(version string, allowNew bool) (*Version, bool) {
	v, ok := a.Versions[version]
	if !ok {
		if allowNew {
			log.Logger.Debugf("GetVersion no data found; returning %+v", v)
			v = &Version{
				Metrics: map[string]*SummaryMetric{},
			}
			a.Versions[version] = v
			return v, true
		}
		return nil, false
	}

	log.Logger.Debugf("GetVersion returning %+v", v)
	return v, false
}

// UnmarchalJSON unmarshals an application from a byte array. This is a
// custom JSON unmarshaller to ensurer that maps are initialized
func (a *Application) UnmarshalJSON(data []byte) error {
	// use type alias to avoid infinite loop
	type Alias Application
	aux := &struct{ *Alias }{Alias: (*Alias)(a)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// initialize a.Versions if not already
	if a.Versions == nil {
		a.Versions = Versions{}
	}
	for _, v := range a.Versions {
		if v.Metrics == nil {
			v.Metrics = map[string]*SummaryMetric{}
		}
	}

	return nil
}

func (a *Application) MarshalJSON() ([]byte, error) {
	// use type Alias to avoid inifinite loop
	type Alias Application
	rawData, err := json.Marshal(&struct{ *Alias }{Alias: (*Alias)(a)})
	if err != nil {
		return rawData, err
	}

	// remove untracked versions if the rawData is too large
	if len(rawData) > maxApplicationDataBytes {
		deleteUntrackedVersions(a)
		rawData, err = json.Marshal(&struct{ *Alias }{Alias: (*Alias)(a)})
		if err != nil {
			return rawData, err
		}
	}

	// if it is still too large, return an error
	if len(rawData) > maxApplicationDataBytes {
		return rawData, errors.New("application data too large")
	}

	return rawData, nil
}

// String returns a string representation of the Application
func (a *Application) String() string {
	tracks := []string{}
	for t, v := range a.Tracks {
		tracks = append(tracks, t+" -> "+v)
	}

	versions := []string{}
	for version := range a.Versions {
		versions = append(versions, version)
	}

	return fmt.Sprintf("Application %s:\n\t%s\n\t%s", a.Name,
		"tracks: ["+strings.Join(tracks, ",")+"]",
		"versions: ["+strings.Join(versions, ",")+"]",
	)
}
