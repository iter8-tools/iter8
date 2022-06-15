package watcher

// watchedobject.go - methods to read fields from an unstructured Kubernetes object

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	NAME_LABEL       = "app.kubernetes.io/name"
	VERSION_LABEL    = "app.kubernetes.io/version"
	READY_ANNOTATION = "iter8.tools/ready"
	TRACK_ANNOTATION = "iter8.tools/track"
)

// WatchedObject is wrapper for object returned by informer
type WatchedObject struct {
	// Obj the object
	Obj *unstructured.Unstructured
}

// getName gets application name from NAME_LABEL label on watched object
func (wo WatchedObject) getName() (string, bool) {
	labels := wo.Obj.GetLabels()
	name, ok := labels[NAME_LABEL]
	return name, ok
}

// getNamespace gets namespace of watched object
func (wo WatchedObject) getNamespace() string {
	return wo.Obj.GetNamespace()
}

// getNamespacedName returns formatted namespace and (application) name
func (wo WatchedObject) getNamespacedName() (string, bool) {
	n, ok := wo.getName()
	return wo.getNamespace() + "/" + n, ok
}

// getVersion gets application version from VERSION_LABEL label on watched object
func (wo WatchedObject) getVersion() (string, bool) {
	labels := wo.Obj.GetLabels()
	v, ok := labels[VERSION_LABEL]
	return v, ok
}

// getTrack get trace of version from TRACK_ANNOTATION annotation on watched object
func (wo WatchedObject) getTrack() string {
	annotations := wo.Obj.GetAnnotations()
	track, ok := annotations[TRACK_ANNOTATION]
	if !ok {
		return ""
	}
	return track
}

// isReady determines if watched object indicates readiness of version (as indicated by READY_ANNOTATION annotatation)
func (wo WatchedObject) isReady(currentlyReady bool) bool {
	annotations := wo.Obj.GetAnnotations()
	ready, ok := annotations[READY_ANNOTATION]
	if !ok {
		return currentlyReady || false
	}
	return strings.ToLower(ready) == "true"
}
