package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	// secretPostfix is the postfix added to an application name to create a secret name
	secretPostfix string = ".iter8abnmetrics"
	// secretKey is the name of the key in the data field of a kubernetes secret in which the application will be written
	secretKey string = "application.yaml"
	// defaultBatchWriteInterval is the default value of BatchWriteInterval
	defaultBatchWriteInterval time.Duration = time.Duration(60 * time.Second)
	// defaultMaxApplicationDataBytes is the default of maxApplicationDataBytes
	// must be less than the maximum size of a Kubernetes secret (1 MB)
	// The size of an application is proportional to the number of versions and the number of metrics per version.
	// Since only summary metrics are permitted, each is a fixed size
	defaultMaxApplicationDataBytes int = 750000
	// defaultFlushMultiplier is the default multiplier used to compute the
	// interval for flushing applications that have not been written in a while
	defaultFlushMultiplier int = 5
)

var (
	// Applications is map of app name to Application
	// This is a global variable used to maintain an internal representation of the applications in a cluster
	Applications ThreadSafeApplicationMap
	// BatchWriteInterval is the interval over which changes may batched before being persisted
	BatchWriteInterval time.Duration
	// flushMultiplier is the interval at which applications are checked to see if they were not persisted because of
	// write calls stopped or are too infrequent.
	flushMultiplier int
	// maxApplicationDataBytes is the maximum number of bytes allowed in an applicaton (as YAML converted to []byte)
	// this limit prevents trying to persist an application that is too large (Kubernetes secrets have a 1 MB size limit)
	maxApplicationDataBytes int
)

// initalize global variables
func init() {
	Applications = ThreadSafeApplicationMap{
		apps:           map[string]*Application{},
		mutexes:        map[string]*sync.RWMutex{},
		lastWriteTimes: map[string]*time.Time{},
	}
	BatchWriteInterval = defaultBatchWriteInterval
	flushMultiplier = defaultFlushMultiplier
	maxApplicationDataBytes = defaultMaxApplicationDataBytes // a secret's maximum size is 1MB
}

// ThreadSafeApplicationMap is type to control thread safety of operations on an application map
type ThreadSafeApplicationMap struct {
	// mutex is used to mediate read/write of the application map (ie, Get vs Add/Clear)
	mutex sync.RWMutex
	apps  map[string]*Application
	// mutexes mediate read/write of individual applications within the map
	mutexes        map[string]*sync.RWMutex
	lastWriteTimes map[string]*time.Time
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

// Get application object.
func (m *ThreadSafeApplicationMap) Get(application string) (*Application, error) {
	// if available in the in-memory map, return it
	m.mutex.RLock()
	a, ok := m.apps[application]
	m.mutex.RUnlock()
	if ok {
		return a, nil
	}

	// otherwise, read from persistent store (secret)
	// if no secret, create new object
	a, err := m.readFromSecret(application)

	// and add to the in memory map
	m.Add(a)

	return a, err
}

// readFromSecret reads the application from persistent storage (a Kubernetes secret)
// - the secret name/namespace is the same as the application
// - if no application is present in the persistent storage, a new object is created
func (m *ThreadSafeApplicationMap) readFromSecret(application string) (*Application, error) {
	secretNamespace := namespaceFromKey(application)
	secretName := nameFromKey(application) + secretPostfix

	newApplication := &Application{
		Name:     application,
		Versions: Versions{},
		Tracks:   Tracks{},
	}

	// read secret from cluster; extract appData
	secret, err := k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Logger.Debug("no secret backing " + application)
		return newApplication, err
	}

	// read data from secret (is a yaml file)
	rawData, ok := secret.Data[secretKey]
	if !ok {
		log.Logger.Debug("key missing in backing secret")
		return newApplication, errors.New("secret does not contain expected key: " + secretKey)
	}

	a := &Application{}
	err = yaml.Unmarshal(rawData, a)
	if err != nil {
		log.Logger.Debug("unmarshal failure")
		return newApplication, nil
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
	// note that this uses a custom MarshalJSON that removes untracked
	// versions if the application data is too large
	rawData, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	secretNamespace := namespaceFromKey(a.Name)
	secretName := secretNameFromKey(a.Name)

	// determine if need to
	exists := true
	secret, err = k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
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

	secret.Data[secretKey] = rawData
	if secret.StringData != nil {
		secret.StringData[secretKey] = string(rawData)
	}

	// create or update the secret
	if exists {
		// TBD do we need to merge what we have?
		_, err = k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Update(
			context.Background(),
			secret,
			metav1.UpdateOptions{},
		)
	} else {
		_, err = k8sclient.Client.Typed().CoreV1().Secrets(secretNamespace).Create(
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

// BatchedWrite writes the Application to persistent storage only if the previous write
// was more than BatchWriteInterval ago. If no more writes take place, it is possible that
// some data is not persisted. To avoid this, the A/B/n service should periodically Flush
// application data.
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

func deleteUntrackedVersions(a *Application) {
	toDelete := []string{}
	for version, v := range a.Versions {
		if v.GetTrack() == nil {
			toDelete = append(toDelete, version)
		}
	}

	for _, version := range toDelete {
		delete(a.Versions, version)
	}
}

// nameFromKey returns the name from a key of the form "namespace/name"
func nameFromKey(applicationKey string) string {
	_, n := splitApplicationKey(applicationKey)
	return n
}

// secretNameFromKey returns the name of the secret used to persist an applicatiob
func secretNameFromKey(applicationKey string) string {
	return nameFromKey(applicationKey) + secretPostfix
}

// namespaceFromKey returns the namespace from a key of the form "namespace/name"
func namespaceFromKey(applicationKey string) string {
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

// PeriodicApplicationsFlush periodically checks if there is any (metric) data associated with
// an application that has not been persisted to the underlying secret. If so, it is written.
// This supports the edge case of an application that stops receiving requests to write metric data.
// The period on which flush works is a multiple of the BatchWriteInterval; it is expected that
// BatchWrite will handle the majority of the required persistence.
func (m *ThreadSafeApplicationMap) PeriodicApplicationsFlush(done chan struct{}) {
	ticker := time.NewTicker(time.Duration(flushMultiplier) * BatchWriteInterval)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				m.flush()
			}
		}
	}()
}

// flush writes all applications whose last write time is greater than the BatchWriteInterval
func (m *ThreadSafeApplicationMap) flush() {
	// get list of applications that need flushing
	now := time.Now()
	toFlush := []string{}
	m.mutex.RLock()
	for application, last := range m.lastWriteTimes {
		if now.Sub(*last) > BatchWriteInterval {
			toFlush = append(toFlush, application)
		}
	}
	m.mutex.RUnlock()

	// flush them .. unless they have been written since we inspected them above
	for _, application := range toFlush {
		a, err := m.Get(application)
		if err != nil {
			continue
		}
		m.BatchedWrite(a)
	}
}
