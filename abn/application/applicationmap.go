package application

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	SECRET_POSTFIX string = ".iter8abnmetrics"
)

var (
	// Applications is map of app name to Application
	// This is a global variable used to maintain an internal representation of the applications in a cluster
	Applications ThreadSafeApplicationMap
	// batchWriteInterval is the interval during which a write may not take place
	BatchWriteInterval time.Duration
)

// initalize global variables
func init() {
	Applications = ThreadSafeApplicationMap{
		apps:           map[string]*Application{},
		mutexes:        map[string]*sync.RWMutex{},
		lastWriteTimes: map[string]*time.Time{},
	}
	BatchWriteInterval = time.Duration(60 * time.Second)
}

// ThreadSafeApplicationMap is type to control thread safety of operations on an application map
type ThreadSafeApplicationMap struct {
	// mutex is used to mediate read/write of the application map (ie, Get vs Add/Clear)
	mutex sync.RWMutex
	apps  map[string]*Application
	// mutexes mediate read/write of individual applications within the map
	mutexes        map[string]*sync.RWMutex
	lastWriteTimes map[string]*time.Time
	rw             *ApplicationReaderWriter
}

// RLock lock application for reading
func (m *ThreadSafeApplicationMap) RLock(application string) {
	m.mutexes[application].RLock()
}

// RUnlock undoes a single RLock call
func (m *ThreadSafeApplicationMap) RUnlock(application string) {
	m.mutexes[application].RUnlock()
}

// Lock lock application for writing
func (m *ThreadSafeApplicationMap) Lock(application string) {
	m.mutexes[application].Lock()
}

// Unlock unlocks application
func (m *ThreadSafeApplicationMap) Unlock(application string) {
	m.mutexes[application].Unlock()
}

// Add adds an application into the application map
func (m *ThreadSafeApplicationMap) Add(a *Application) {
	m.mutex.Lock()
	m.apps[a.Name] = a
	m.mutexes[a.Name] = &sync.RWMutex{}
	m.mutex.Unlock()
}

// SetReaderWriter sets the ReaderWriter (for reading/writing secrets to a cluster)
func (m *ThreadSafeApplicationMap) SetReaderWriter(rw *ApplicationReaderWriter) {
	m.rw = rw
}

// Get gets an application from map of applications
// If the application is not present and a reader is provided, an attempt will be made to
// read it from persistant storage
func (m *ThreadSafeApplicationMap) Get(application string, inMemoryOnly bool) (*Application, error) {
	m.mutex.RLock()
	a, ok := m.apps[application]
	m.mutex.RUnlock()
	if ok {
		return a, nil
	}

	// !ok
	if m.rw == nil || inMemoryOnly {
		return nil, errors.New("application record not found in memory")
	}

	a, err := m.Read(application)
	m.Add(a)
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

	// set last write time to read time; it was written in the past
	now := time.Now()
	m.lastWriteTimes[a.Name] = &now

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
		return err
	}

	// update last write time for application
	now := time.Now()
	m.lastWriteTimes[a.Name] = &now
	return nil
}

// BatchedWrite writes the Application to persistent storage only if the previous write was more than TIME ago
func (m *ThreadSafeApplicationMap) BatchedWrite(a *Application) error {
	log.Logger.Tracef("BatchedWrite called")
	defer log.Logger.Trace("BatchedWrite completed")

	now := time.Now()
	lastWrite, ok := m.lastWriteTimes[a.Name]
	if !ok || lastWrite == nil {
		// no record of the application ever being written; write it now
		m.Write(a)
	} else {
		if now.Sub(*m.lastWriteTimes[a.Name]) > BatchWriteInterval {
			m.Write(a)
		}
	}

	// it was written too recently; wait until another write call
	return nil
}
