package watcher

// app.go - methods to track applications and their versions

import (
	"github.com/iter8-tools/iter8/base/log"
)

// Application is information about versions of an application
type Application struct {
	// versions is map of versions for this application
	versions map[string]Version
	// map of tracks to version
	tracks map[string]string
}

// Version is version of an Application
type Version struct {
	// Name of version
	Name string
	// Ready indicates if version is ready
	Ready bool
	// Track is track of version
	Track string
}

// apps is map of app name to Application
var apps map[string]Application = map[string]Application{}

// Add updates the apps map using information from a newly added object
func Add(watched WatchedObject) {
	log.Logger.Trace("Add called")
	defer log.Logger.Trace("Add completed")
	// Assume applications are namespace scoped; use name in form: "namespace/name"
	// where name is the value of the label app.kubernetes.io/name
	name, ok := watched.getNamespacedName()
	if !ok {
		// no name; ignore the object
		log.Logger.Trace("no name found")
		return
	}

	// Expect version using labe app.kubernetes.io/version
	version, ok := watched.getVersion()
	if !ok {
		// no version; ignore the object
		log.Logger.Debug("no version found")
		return
	}

	// check if we know about this application; if not create entry
	app, ok := apps[name]
	if !ok {
		// create record of discovered app if not already present
		app = Application{
			versions: map[string]Version{},
			tracks:   map[string]string{},
		}
		// record new application
		apps[name] = app
	}

	// check if we know about this version; if not create entry
	v, ok := apps[name].versions[version]
	if !ok {
		// create record of discovered version
		v = Version{
			Name:  version,
			Ready: false,
		}
	}

	// set ready to value on watched object, if set
	// otherwise, use the current readiness value
	v.Ready = watched.isReady(v.Ready)

	// update track <--> ready version mapping
	if v.Ready {
		watchedTrack := watched.getTrack()
		if watchedTrack != "" {
			// update track for version
			v.Track = watchedTrack
			// update version for track
			app.tracks[watchedTrack] = v.Name
		}
	} else {
		// if version has track then unmap it
		// but first check the track to version and remove if mapped to this (not ready) version
		if v.Track != "" {
			_, ok := app.tracks[v.Track]
			if ok {
				delete(app.tracks, v.Track)
			}
		}
		// v not ready, remove any map to track
		v.Track = ""
	}

	// record update
	app.versions[version] = v
}

// Update updates the apps map using information from a modified object
func Update(watched WatchedObject) {
	log.Logger.Trace("Update called")
	defer log.Logger.Trace("Update completed")

	Add(watched)
}

// Delete updates the apps map using information from a deleted object
func Delete(watched WatchedObject) {
	log.Logger.Trace("Delete called")
	defer log.Logger.Trace("Delete called")

	name, ok := watched.getNamespacedName()
	if !ok {
		return // no app.kubernetes.io/name label
	}
	_, ok = apps[name]
	if !ok {
		return // has app.kubernetes.io/name but object wasn't recorded
	}

	version, ok := watched.getVersion()
	if !ok {
		return // no app.kubernetes.io/version label
	}
	v, ok := apps[name].versions[version]
	if !ok {
		return // no version recorded (should not happen)
	}

	// if object being deleted has ready annotation we are no longer ready
	annotations := watched.Obj.GetAnnotations()
	if _, ok := annotations[READY_ANNOTATION]; ok {
		v.Ready = false
		_, ok := apps[name].tracks[v.Track]
		if ok {
			delete(apps[name].tracks, v.Track)
		}
		v.Track = ""
	}

	apps[name].versions[version] = v

	if len(apps[name].versions) == 0 {
		delete(apps, name)
	}
}

// dump logs the apps map
// used for debug only
func dump() {
	for name, app := range apps {
		log.Logger.Tracef("\nAPPLICATION: %s\n", name)
		for version, v := range app.versions {
			log.Logger.Tracef(" > version = %s, track = %s, ready = %t\n", version, v.Track, v.Ready)
		}
		for t, v := range app.tracks {
			log.Logger.Tracef(" > track = %s, version = %s", t, v)
		}
	}
}
