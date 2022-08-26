package application

// application.go - type of application

import (
	"fmt"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/client-go/kubernetes"
)

const (
	// KEY is the name of the field in a kubernetes secret
	KEY = "versionData.yaml"
)

// Application is an application observed in a kubernetes cluster
type Application struct {
	// Name is the value of the app.kubernetes.io/name field
	Name string
	// Namespace is the namespace where the application was discovered
	Namespace string
	// Tracks is map from application track identifier to version name
	Tracks
	// Versions is a map of versions name to version data
	Versions
}

// Versions is type of version name to versions map
type Versions map[string]*Version

// Tracks is map of track identifiers to version names
type Tracks map[string]string

// ApplicationReaderWriter is type used to read/write fromt/to persistent storage
type ApplicationReaderWriter struct {
	// Client is the Kubernetes client to use to read/write secrets
	Client kubernetes.Interface
}

// NewApplication returns a new (empty) Application for a namespace/name label
func NewApplication(application string) *Application {
	return &Application{
		Name:      GetNameFromKey(application),
		Namespace: GetNamespaceFromKey(application),
		Versions:  Versions{},
		Tracks:    Tracks{},
	}
}

// GetNameFromKey returns the name from a key of the form "namespace/name"
func GetNameFromKey(applicationKey string) string {
	_, n := splitApplicationKey(applicationKey)
	return n
}

// GetNamespaceFromKey returns the namespace from a key of the form "namespace/name"
func GetNamespaceFromKey(applicationKey string) string {
	ns, _ := splitApplicationKey(applicationKey)
	return ns
}

// splitApplicationKey is a utility function that returns the name and namespace from a key of the form "namespace/name"
func splitApplicationKey(applicationKey string) (string, string) {
	var name, namespace string
	names := strings.Split(applicationKey, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return namespace, name
}

// GetVersion returns the Version identified by version
// when allowNew is true, a new (blank) Version will  be created if none can be found
// returns the version and a boolean indicating whether or not a new version was created or not
func (a *Application) GetVersion(version string, allowNew bool) (*Version, bool) {
	v, ok := a.Versions[version]
	if !ok {
		if allowNew {
			log.Logger.Debugf("GetVersion no data found; returning %+v", v)
			v = &Version{
				// History:             []VersionEvent{},
				Metrics:             map[string]*SummaryMetric{},
				LastUpdateTimestamp: time.Now(),
			}
			a.Versions[version] = v
			return v, true
		}
		return nil, true
	}

	log.Logger.Debugf("GetVersion returning %+v", v)
	return v, false
}

// String returns a string representation of the Application
func (a *Application) String() string {
	tracks := []string{}
	for t, v := range a.Tracks {
		tracks = append(tracks, t+" -> "+v)
	}

	str := fmt.Sprintf("Application %s/%s:\n\t%s", a.Namespace, a.Name,
		"tracks: ["+strings.Join(tracks, ",")+"]")

	for version, v := range a.Versions {
		str += fmt.Sprintf("\n\tversion %s%s", version, v)
	}

	return str
}
