package watcher

import (
	"github.com/iter8-tools/iter8/base/log"

	"stathat.com/c/consistent"
)

// Application is information about versions of an application
type Application struct {
	// versions is map of versions for this application
	versions map[string]Version
	// c is holds information for members (versions) of consistent hash circle
	c *consistent.Consistent
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

var apps map[string]Application = map[string]Application{}

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
			c:        consistent.New(),
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
	// if readiness changed, add/remove to/from consistent hash circle
	if v.Ready {
		apps[name].c.Add(v.Name)
	} else {
		apps[name].c.Remove(v.Name)
	}

	// set track if not already set; once set this does not change
	if len(v.Track) == 0 {
		v.Track = watched.getTrack()
	}

	// record update
	app.versions[version] = v
}

func Update(watched WatchedObject) {
	log.Logger.Trace("Update called")
	defer log.Logger.Trace("Update completed")

	Add(watched)
}

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
		apps[name].c.Remove(v.Name)
	}

	apps[name].versions[version] = v

	if len(apps[name].versions) == 0 {
		delete(apps, name)
	}
}

// for debug only
func dump() {
	for name, app := range apps {
		log.Logger.Tracef("application: %s\n", name)
		for version, v := range app.versions {
			readyMsg := "ready"
			if !v.Ready {
				readyMsg = "not " + readyMsg
			}
			log.Logger.Tracef("   version: %s (%s): %s\n", version, v.Track, readyMsg)
		}
	}
}
