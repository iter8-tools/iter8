package watcher

// app.go - methods to track the runtime state of applications and their versions
// For each version, maintain information about its readiness and its mapping to a "track", if any.

// A track is a (user assigned) identifier that the user assigns to versions as part of the CI/CD process.
// When Iter8 A/N(/n) service is used to lookup versions, the track identifier is returned.
// The caller can use this to route requests to the appropriate version.
// To do this, the set of track identifiers should be a (small) fixed set, such as "current" and
// "candidate", that can be mapped to a static routes.

import (
	"github.com/iter8-tools/iter8/abn/appsummary"
	"github.com/iter8-tools/iter8/base/log"
)

// Application is runtime information about the versions of an application
type Application struct {
	// Versions is map of versions for this application
	Versions map[string]Version
	// Tracks maps of track identifiers to versions for quick lookup
	Tracks map[string]string
	// Recorder used to persist events and metrics (to an ApplicationSummary)
	Recorder *appsummary.MetricDriver
}

// Version is runtime details of a version of an Application
type Version struct {
	// Name of version
	Name string
	// Ready indicates if version is ready
	Ready bool
	// Track is track of version
	Track string
}

// Apps is map of app name to Application
var Apps map[string]Application = map[string]Application{}

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

	// check if we know about this application; if not create entry
	app, ok := Apps[name]
	if !ok {
		// create record of discovered app if not already present
		app = Application{
			Versions: map[string]Version{},
			Tracks:   map[string]string{},
		}
		recorder := appsummary.MetricDriver{Client: watched.Driver.Clientset}
		app.Recorder = &recorder

		// record new application
		Apps[name] = app
	}

	recorder := Apps[name].Recorder

	// check if we know about this version; if not create entry
	v, ok := Apps[name].Versions[version]
	if !ok {
		// create record of discovered version
		v = Version{
			Name:  version,
			Ready: false,
		}
		// log new version identified
		recorder.RecordEvent(name, version, appsummary.VersionNewEvent)
	}

	// set ready to value on watched object, if set
	// otherwise, use the current readiness value
	wasReady := v.Ready
	v.Ready = watched.isReady(v.Ready)

	// update track <--> ready version mapping
	if v.Ready {
		// log version ready (if it wasn't before)
		if !wasReady {
			recorder.RecordEvent(name, version, appsummary.VersionReadyEvent)
		}
		watchedTrack := watched.getTrack()
		if watchedTrack != "" {
			oldTrack := v.Track
			// update track for version
			v.Track = watchedTrack
			// update version for track
			app.Tracks[watchedTrack] = v.Name

			// log maptrack event if mapped to a new track
			if oldTrack != v.Track {
				recorder.RecordEvent(name, version, appsummary.VersionMapTrackEvent, v.Track)
			}
		}
	} else {
		// version not ready so if version has track then unmap it
		// but first check the track to version and remove if mapped to this (not ready) version
		if v.Track != "" {
			delete(app.Tracks, v.Track)
			// log unmaptrack event
			recorder.RecordEvent(name, version, appsummary.VersionUnmapTrackEvent)
		}
		// v not ready, remove any map to track
		v.Track = ""
	}

	// record update into Apps
	Apps[name].Versions[version] = v
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
	_, ok = Apps[name]
	if !ok {
		return // has app.kubernetes.io/name but object wasn't recorded
	}

	recorder := Apps[name].Recorder

	version, ok := watched.getVersion()
	if !ok {
		return // no app.kubernetes.io/version label
	}
	v, ok := Apps[name].Versions[version]
	if !ok {
		return // no version recorded (should not happen)
	}

	// if object being deleted has ready annotation we are no longer ready
	annotations := watched.Obj.GetAnnotations()
	if _, ok := annotations[READY_ANNOTATION]; ok {
		// it was ready; record that it is no longer ready
		if v.Ready {
			recorder.RecordEvent(name, version, appsummary.VersionNoLongerReadyEvent)
		}
		v.Ready = false

		// if it was mapped to a track; mark it unmapped
		_, ok := Apps[name].Tracks[v.Track]
		if ok {
			recorder.RecordEvent(name, version, appsummary.VersionUnmapTrackEvent)
		}
		delete(Apps[name].Tracks, v.Track)
		v.Track = ""
	}

	Apps[name].Versions[version] = v

	if len(Apps[name].Versions) == 0 {
		delete(Apps, name)
	}
}
