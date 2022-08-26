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

// // Read reads the application from persistent storage (a Kubernetes secret)
// // - the secret name/namespace is the same as the application
// // - if no application is present in the persistent storage, a new object is created
// func (rw *ApplicationReaderWriter) Read(appName string) (*Application, error) {
// 	a := NewApplication(appName, rw)

// 	// read secret from cluster; extract appData
// 	secret, err := rw.Client.CoreV1().Secrets(a.Namespace).Get(context.Background(), a.Name, metav1.GetOptions{})
// 	if err != nil {
// 		log.Logger.Debug("no secret backing " + a.Name)
// 		return a, err
// 	}

// 	// read data from secret (is a yaml file)
// 	rawData, ok := secret.Data[KEY]
// 	if !ok {
// 		log.Logger.Debug("key missing in backing secret")
// 		return a, errors.New("secret does not contain expected key: " + KEY)
// 	}

// 	err = yaml.Unmarshal(rawData, &a.Versions)
// 	if err != nil {
// 		log.Logger.Debug("unmarshal failure")
// 		return a, nil
// 	}

// 	// initialize Tracks and initialize where unmarshal fails to do so
// 	for version, v := range a.Versions {
// 		track := v.GetTrack()
// 		if track != nil {
// 			a.Tracks[*track] = version
// 		}
// 		if v.History == nil {
// 			v.History = []VersionEvent{}
// 		}
// 		if v.Metrics == nil {
// 			v.Metrics = map[string]*SummaryMetric{}
// 		}
// 	}

// 	return a, nil
// }

// // Write writes the Application to persistent storage (a Kubernetes secret)
// func (a *Application) Write() error {
// 	log.Logger.Tracef("Write called with %#v", a)
// 	defer log.Logger.Trace("Write completed")
// 	var secret *corev1.Secret

// 	// marshal to byte array
// 	rawData, err := yaml.Marshal(a.Versions)
// 	if err != nil {
// 		return err
// 	}

// 	// determine if need to
// 	exists := true
// 	secret, err = a.ReaderWriter.Client.CoreV1().Secrets(a.Namespace).Get(context.Background(), a.Name, metav1.GetOptions{})
// 	if err != nil {
// 		exists = false
// 		secret = &corev1.Secret{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      a.Name,
// 				Namespace: a.Namespace,
// 			},
// 			Data:       map[string][]byte{},
// 			StringData: map[string]string{},
// 		}
// 		log.Logger.Debug("secret does not exist; creating")
// 	}

// 	secret.Data[KEY] = rawData
// 	if secret.StringData != nil {
// 		secret.StringData[KEY] = string(rawData)
// 	}

// 	// create or update the secret
// 	if exists {
// 		// TBD do we need to merge what we have?
// 		_, err = a.ReaderWriter.Client.CoreV1().Secrets(a.Namespace).Update(
// 			context.Background(),
// 			secret,
// 			metav1.UpdateOptions{},
// 		)
// 	} else {
// 		_, err = a.ReaderWriter.Client.CoreV1().Secrets(a.Namespace).Create(
// 			context.Background(),
// 			secret,
// 			metav1.CreateOptions{},
// 		)
// 	}
// 	if err != nil {
// 		log.Logger.WithError(err).Warn("unable to persist application")
// 	}

// 	return err

// }

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
