package application

// application.go - type of application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
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
	// Tracks is map of tracks identified to versions
	Tracks
	// Versions is list of discovered versions
	Versions
	// ReaderWriter is used to read/write to persistent storage (a Kubernetes secret)
	Writer *ApplicationReaderWriter
}

// Versions is type of version name to versions map
type Versions map[string]*Version

// Tracks is map of track identifiers to version names
type Tracks map[string]string

// ApplicationReaderWriter is type used to read/write fromt/to persistent storage
type ApplicationReaderWriter struct {
	// Type Kubernetes client
	Client kubernetes.Interface
}

// Read reads the application from persistent storage (a Kubernetes secret)
// - the secret name/namespace is the same as the application
// - if no application is present in the persistent storage, a new object is created
func (rw *ApplicationReaderWriter) Read(appName string) (*Application, error) {
	a := GetNewApplication(appName, rw)

	// read secret from cluster; extract appData
	secret, err := rw.Client.CoreV1().Secrets(a.Namespace).Get(context.Background(), a.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger.Debug("no secret backing " + a.Name)
		return a, err
	}

	// read data from secret (is a yaml file)
	rawData, ok := secret.Data[KEY]
	if !ok {
		log.Logger.Debug("key missing in backing secret")
		return a, errors.New("secret does not contain expected key: " + KEY)
	}

	var versions Versions
	err = yaml.Unmarshal(rawData, &versions)
	if err != nil {
		log.Logger.Debug("unmarshal failure")
		return a, nil
	}

	// set Versions
	a.Versions = versions

	// initialize Tracks
	for version, v := range versions {
		track := v.GetTrack()
		if track != nil {
			a.Tracks[*track] = version
		}
		if v.History == nil {
			v.History = []VersionEvent{}
		}
		if v.Metrics == nil {
			v.Metrics = map[string]*SummaryMetric{}
		}
	}

	return a, nil
}

// Write writes the Application to persistent storage (a Kubernetes secret)
func (a *Application) Write() error {
	log.Logger.Tracef("Write called with %#v", a)
	defer log.Logger.Trace("Write completed")
	var secret *corev1.Secret

	// marshal to byte array
	rawData, err := yaml.Marshal(a.Versions)
	if err != nil {
		return err
	}

	// determine if need to
	exists := true
	secret, err = a.Writer.Client.CoreV1().Secrets(a.Namespace).Get(context.Background(), a.Name, metav1.GetOptions{})
	if err != nil {
		exists = false
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      a.Name,
				Namespace: a.Namespace,
			},
			Data:       map[string][]byte{},
			StringData: map[string]string{},
		}
		log.Logger.Debug("secret does not exist; creating")
	}

	secret.Data[KEY] = rawData
	if secret.StringData != nil {
		secret.StringData[KEY] = string(rawData)
	}

	// create or update the secret
	if exists {
		// TBD do we need to merge what we have?
		_, err = a.Writer.Client.CoreV1().Secrets(a.Namespace).Update(
			context.Background(),
			secret,
			metav1.UpdateOptions{},
		)
	} else {
		_, err = a.Writer.Client.CoreV1().Secrets(a.Namespace).Create(
			context.Background(),
			secret,
			metav1.CreateOptions{},
		)
	}
	if err != nil {
		log.Logger.WithError(err).Warn("unable to persist application")
	}

	return err

}

// GetNewApplication returns a new (empty) Application for a namespace/name label
func GetNewApplication(application string, rw *ApplicationReaderWriter) *Application {
	var name, namespace string
	names := strings.Split(application, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	a := Application{
		Name:      name,
		Namespace: namespace,
		Versions:  Versions{},
		Tracks:    Tracks{},
		Writer:    rw,
	}

	return &a
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
				History:             []VersionEvent{},
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
