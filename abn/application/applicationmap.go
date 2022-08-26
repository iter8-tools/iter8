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

const (
	SECRET_POSTFIX string = ".iter8abnmetrics"
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
	m.Lock()
	m.apps = map[string]*Application{}
	m.Unlock()
}

func (m *ThreadSafeApplicationMap) Add(key string, a *Application) {
	m.Lock()
	m.apps[key] = a
	m.Unlock()
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
		return nil, errors.New("application record not found in memory")
	}

	a, err := m.Read(application)
	Applications.apps[application] = a
	return a, err
}

// Read reads the application from persistent storage (a Kubernetes secret)
// - the secret name/namespace is the same as the application
// - if no application is present in the persistent storage, a new object is created
func (m *ThreadSafeApplicationMap) Read(application string) (*Application, error) {
	secretNamespace := GetNamespaceFromKey(application)
	secretName := GetNameFromKey(application) + SECRET_POSTFIX

	// read secret from cluster; extract appData
	secret, err := m.rw.Client.CoreV1().Secrets(secretNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Logger.Debug("no secret backing " + application)
		return NewApplication(application), err
	}

	// read data from secret (is a yaml file)
	rawData, ok := secret.Data[KEY]
	if !ok {
		log.Logger.Debug("key missing in backing secret")
		return NewApplication(application), errors.New("secret does not contain expected key: " + KEY)
	}

	// err = yaml.Unmarshal(rawData, &a.Versions)
	a := &Application{}
	err = yaml.Unmarshal(rawData, a)
	if err != nil {
		log.Logger.Debug("unmarshal failure")
		return NewApplication(application), nil
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

	return a, nil
}

// Write writes the Application to persistent storage (a Kubernetes secret)
func (m *ThreadSafeApplicationMap) Write(a *Application) error {
	log.Logger.Tracef("Write called with %#v", a)
	defer log.Logger.Trace("Write completed")
	var secret *corev1.Secret

	// marshal to byte array
	rawData, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	secretNamespace := GetNamespaceFromKey(a.Name)
	secretName := GetSecretNameFromKey(a.Name)

	// determine if need to
	exists := true
	secret, err = m.rw.Client.CoreV1().Secrets(secretNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		exists = false
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: secretNamespace,
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
		_, err = m.rw.Client.CoreV1().Secrets(secretNamespace).Update(
			context.Background(),
			secret,
			metav1.UpdateOptions{},
		)
	} else {
		_, err = m.rw.Client.CoreV1().Secrets(secretNamespace).Create(
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