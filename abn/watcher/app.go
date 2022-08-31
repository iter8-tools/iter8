package watcher

// app.go - methods to track the runtime state of applications and their versions
// For each version, maintain information about its readiness and its mapping to a "track", if any.

// A track is a (user assigned) identifier that the user assigns to versions as part of the CI/CD process.
// When Iter8 A/N(/n) service is used to lookup versions, the track identifier is returned.
// The caller can use this to route requests to the appropriate version.
// To do this, the set of track identifiers should be a (small) fixed set, such as "current" and
// "candidate", that can be mapped to a static routes.

import (
	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
)

// addObject updates the apps map using information from a newly added object
// If the observed object does not have a name (app.kubernetes.io/name label)
// or version (app.kubenetes.io/version), it is ignored.
func addObject(watched WatchedObject) {
	log.Logger.Tracef("Add called for %s/%s", watched.Obj.GetNamespace(), watched.Obj.GetName())
	defer log.Logger.Trace("Add completed")

	// Is the object involved in a ABn experiment?
	if !watched.isIter8AbnRelated() {
		log.Logger.Debug("not Iter8 abn related")
		return
	}

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

	// the watched object is the object that defines the version of the application

	// check if we know about this application
	// first check if in memory then read from persistent store if not found
	// if it isn't in persistent store, the read will return an initalized Application
	a, _ := abnapp.Applications.Get(name, false)

	abnapp.Applications.Lock(name)
	defer abnapp.Applications.Unlock(name)

	// get the version; if it isn't in the Application this will create an new Version
	v, _ := a.GetVersion(version, true)

	// update track <--> version mapping
	if !watched.isReady() {
		// version not ready; ensure no track is nil and not in a.Tracks
		if v.Track != nil {
			delete(a.Tracks, *v.Track)
		}
	} else {
		// version is ready; set v.Track and add to a.Tracks if defined
		watchedTrack := watched.getTrack()
		if watchedTrack == "" {
			// track not set; ensure not in a.Tracks and ensure is nil
			if v.Track != nil {
				delete(a.Tracks, *v.Track)
			}
			v.Track = nil
		} else {
			// track is set
			v.Track = &watchedTrack
			a.Tracks[*v.Track] = version
		}
	}

	// record update into Apps
	err := abnapp.Applications.Write(a)
	if err != nil {
		log.Logger.Error("unable to write application")
	}
}

// updateObject updates the apps map using information from a modified object
// Behavior is the same as for a new object
func updateObject(watched WatchedObject) {
	log.Logger.Trace("Update called")
	defer log.Logger.Trace("Update completed")

	addObject(watched)
}

// deleteObject updates the apps map using information from a deleted object
// Note that we are not object counting which means we will never actually remove a version
// from an application or an application from the syste
func deleteObject(watched WatchedObject) {
	log.Logger.Trace("Delete called")
	defer log.Logger.Trace("Delete called")

	name, ok := watched.getNamespacedName()
	if !ok {
		return // no app.kubernetes.io/name label
	}

	_, err := abnapp.Applications.Get(name, false)
	if err != nil {
		return // has app.kubernetes.io/name but object wasn't recorded
	}

	version, ok := watched.getVersion()
	if !ok {
		return // no app.kubernetes.io/version label
	}

	a, _ := abnapp.Applications.Get(name, true)
	if a == nil {
		return // no record; we don't look in secret if we got a delete event, we must have had an add/update event
	}

	abnapp.Applications.Lock(name)
	defer abnapp.Applications.Unlock(name)

	v, _ := a.GetVersion(version, false)
	if v == nil {
		return // no version was recorded; on delete this should not happen
	}

	// if object being deleted has ready annotation we are no longer ready
	// set track to nil and remove from a.Tracks
	if v.Track != nil {
		delete(a.Tracks, *v.Track)
	}
	v.Track = nil

}
