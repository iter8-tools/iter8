package watcher

// app.go - methods to track the runtime state of applications and their versions
// For each version, maintain information about its readiness and its mapping to a "track", if any.

// A track is a (user assigned) identifier that the user assigns to versions as part of the CI/CD process.
// When Iter8 A/N(/n) service is used to lookup versions, the track identifier is returned.
// The caller can use this to route requests to the appropriate version.
// To do this, the set of track identifiers should be a (small) fixed set, such as "current" and
// "candidate", that can be mapped to a static routes.

import (
	"sync"

	"github.com/iter8-tools/iter8/abn/application"
	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
)

// ThreadSafeApplicationMap is type to control thread safety of operations on an application map
type ThreadSafeApplicationMap struct {
	mutex sync.Mutex
	apps  map[string]*abnapp.Application
}

// Applications is map of app name to Application
// This is a global variable used to maintain an internal representation of the applications in a cluster
var Applications = ThreadSafeApplicationMap{
	apps: map[string]*abnapp.Application{},
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
	m.apps = map[string]*abnapp.Application{}
	m.mutex.Unlock()
}

func (m *ThreadSafeApplicationMap) Add(key string, a *abnapp.Application) {
	m.mutex.Lock()
	m.apps[key] = a
	m.mutex.Unlock()
}

// Get gets an application from map of applications
// If the application is not present and a reader is provided, an attempt will be made to
// read it from persistant storage
// Applications.Lock() should be called first
func (m *ThreadSafeApplicationMap) Get(key string, reader *abnapp.ApplicationReaderWriter) (*abnapp.Application, error) {
	a, ok := Applications.apps[key]
	if !ok {
		if reader == nil {
			return nil, nil
		}
		a, err := reader.Read(key)
		Applications.apps[key] = a
		return a, err
	}
	return a, nil
}

// Add updates the apps map using information from a newly added object
// If the observed object does not have a name (app.kubernetes.io/name label)
// or version (app.kubenetes.io/version), it is ignored.
func Add(watched WatchedObject) {
	log.Logger.Tracef("Add called for %s/%s", watched.Obj.GetNamespace(), watched.Obj.GetName())
	defer log.Logger.Trace("Add completed")

	// Assume applications are namespace scoped; use name in form: "namespace/name"
	// where name is the value of the label app.kubernetes.io/name
	name, ok := watched.getNamespacedName()
	if !ok {
		// no name; ignore the object
		log.Logger.Debug("no name found")
		return
	}

	// Expect version using labe app.kubernetes.io/version
	version, ok := watched.getVersion()
	if !ok {
		// no version; ignore the object
		log.Logger.Debug("no version found")
		return
	}

	Applications.mutex.Lock()
	defer Applications.mutex.Unlock()

	// check if we know about this application
	// first check if in memory
	// if not, read from persistent store
	// if it does not exist in persistent store, the read will return an initalized Application
	a, _ := Applications.Get(name, watched.Writer)

	// get the version
	// if it isn't in the Application this will create an new Version
	v, isNew := a.GetVersion(version, true)
	if isNew {
		v.AddEvent(application.VersionNewEvent)
	}

	// set ready to value on watched object, if set
	// otherwise, use the current readiness value
	oldReady := v.IsReady()
	watchedReady := watched.isReady(oldReady)

	// update track <--> ready version mapping
	if watchedReady {
		// log version ready (if it wasn't before)
		if !oldReady {
			v.AddEvent(application.VersionReadyEvent)
		}
		watchedTrack := watched.getTrack()
		if watchedTrack != "" {
			oldTrack := v.GetTrack()
			// log maptrack event if mapped to a new track
			if oldTrack == nil || *oldTrack != watchedTrack {
				v.AddEvent(application.VersionMapTrackEvent, watchedTrack)
				// update a.Tracks
				a.Tracks[watchedTrack] = version
			}
		}
	} else {
		// version not ready so if version has track then unmap it
		// but first check the track to version and remove if mapped to this (not ready) version
		oldTrack := v.GetTrack()
		if oldTrack != nil {
			delete(a.Tracks, *oldTrack)
			// log unmaptrack event
			v.AddEvent(application.VersionUnmapTrackEvent)
		}
		v.AddEvent(application.VersionNoLongerReadyEvent)
	}

	// record update into Apps
	toWrite := Applications.apps[name]
	err := toWrite.Write()
	if err != nil {
		log.Logger.Error("unable to write application")
	}
}

// Update updates the apps map using information from a modified object
func Update(watched WatchedObject) {
	log.Logger.Trace("Update called")
	defer log.Logger.Trace("Update completed")

	Add(watched)
}

// Delete updates the apps map using information from a deleted object
// Note that we are not object counting which means we will never actually remove a version
// from an application or an application from the syste
func Delete(watched WatchedObject) {
	log.Logger.Trace("Delete called")
	defer log.Logger.Trace("Delete called")

	Applications.Lock()
	defer Applications.Unlock()

	name, ok := watched.getNamespacedName()
	if !ok {
		return // no app.kubernetes.io/name label
	}
	_, ok = Applications.apps[name]
	if !ok {
		return // has app.kubernetes.io/name but object wasn't recorded
	}

	version, ok := watched.getVersion()
	if !ok {
		return // no app.kubernetes.io/version label
	}

	a, _ := Applications.Get(name, nil)
	if a == nil {
		return // no record; we don't look in secret if we got a delete event, we must have had an add/update event
	}

	v, _ := a.GetVersion(version, false)
	if v == nil {
		return // no version was recorded; on delete this should not happen
	}

	// if object being deleted has ready annotation we are no longer ready
	versionReady := v.IsReady()
	watchedReady := watched.isReady(false)
	versionTrack := v.GetTrack()

	if watchedReady {
		// it was ready; record that it is no longer ready
		if versionReady {
			v.AddEvent(application.VersionNoLongerReadyEvent)
		}

		// if it was mapped to a track; mark it unmapped (since no longer ready)
		if versionTrack != nil {
			v.AddEvent(application.VersionUnmapTrackEvent)
			delete(a.Tracks, *versionTrack)
		}
	}

	// Applications[name].Versions[version] = v

	if len(Applications.apps[name].Versions) == 0 {
		delete(Applications.apps, name)
	}
}
