package application

import (
	"context"
	"errors"
	"sync"

	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// ThreadSafeApplicationMap is type to control thread safety of operations on an application map
type ThreadSafeApplicationMap struct {
	mutex sync.Mutex
	apps  map[string]*Application
	rw    *ApplicationReaderWriter
}

// Applications is map of app name to Application
// This is a global variable used to maintain an internal representation of the applications in a cluster
var Applications = ThreadSafeApplicationMap{
	apps: map[string]*Application{},
}

// Lock locks the application map; should always be followed by an Unlock()
func (m *ThreadSafeApplicationMap) Lock() {
	m.mutex.Lock()
}

// Unlock unlocks the application map
func (m *ThreadSafeApplicationMap) Unlock() {
	m.mutex.Unlock()
}

// Clear the application map
func (m *ThreadSafeApplicationMap) Clear() {
	m.mutex.Lock()
	m.apps = map[string]*Application{}
	m.mutex.Unlock()
}

func (m *ThreadSafeApplicationMap) Add(key string, a *Application) {
	m.mutex.Lock()
	m.apps[key] = a
	m.mutex.Unlock()
}

func (m *ThreadSafeApplicationMap) SetReaderWriter(rw *ApplicationReaderWriter) {
	m.rw = rw
}

// Get gets an application from map of applications
// If the application is not present and a reader is provided, an attempt will be made to
// read it from persistant storage
// Applications.Lock() should be called first
func (m *ThreadSafeApplicationMap) Get(application string, inMemoryOnly bool) (*Application, error) {
	a, ok := Applications.apps[application]
	if ok {
		return a, nil
	}

	// !ok
	if m.rw == nil || inMemoryOnly {
		return nil, nil
	}

	a, err := m.Read(application)
	Applications.apps[application] = a
	return a, err
}

// Read reads the application from persistent storage (a Kubernetes secret)
// - the secret name/namespace is the same as the application
// - if no application is present in the persistent storage, a new object is created
func (m *ThreadSafeApplicationMap) Read(application string) (*Application, error) {
	a := NewApplication(application)

	// read secret from cluster; extract appData
	secret, err := m.rw.Client.CoreV1().Secrets(a.Namespace).Get(context.Background(), a.Name, metav1.GetOptions{})
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

	err = yaml.Unmarshal(rawData, &a.Versions)
	if err != nil {
		log.Logger.Debug("unmarshal failure")
		return a, nil
	}

	// initialize Tracks and initialize where unmarshal fails to do so
	for version, v := range a.Versions {
		track := v.GetTrack()
		if track != nil {
			a.Tracks[*track] = version
		}
		if v.Metrics == nil {
			v.Metrics = map[string]*SummaryMetric{}
		}
	}

	return a, nil
}

// Write writes the Application to persistent storage (a Kubernetes secret)
func (m *ThreadSafeApplicationMap) Write(a *Application) error {
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
	secret, err = m.rw.Client.CoreV1().Secrets(a.Namespace).Get(context.Background(), a.Name, metav1.GetOptions{})
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
		_, err = m.rw.Client.CoreV1().Secrets(a.Namespace).Update(
			context.Background(),
			secret,
			metav1.UpdateOptions{},
		)
	} else {
		_, err = m.rw.Client.CoreV1().Secrets(a.Namespace).Create(
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
