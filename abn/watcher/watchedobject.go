package watcher

// watchedobject.go - methods to read fields from an unstructured Kubernetes object

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	nameLabel    = "app.kubernetes.io/name"
	versionLabel = "app.kubernetes.io/version"
	trackLabel   = "iter8.tools/track"
	iter8Label   = "iter8.tools/abn"
)

// watchedObject is wrapper for object returned by informer
type watchedObject struct {
	// Obj is the Kubernetes object
	Obj *unstructured.Unstructured
}

// getName gets application name from NAME_LABEL label on watched object
func (wo watchedObject) getName() (string, bool) {
	labels := wo.Obj.GetLabels()
	name, ok := labels[nameLabel]
	return name, ok
}

// getNamespace gets namespace of watched object
func (wo watchedObject) getNamespace() string {
	return wo.Obj.GetNamespace()
}

// getNamespacedName returns formatted namespace and (application) name
func (wo watchedObject) getNamespacedName() (string, bool) {
	n, ok := wo.getName()
	return wo.getNamespace() + "/" + n, ok
}

// getVersion gets application version from VERSION_LABEL label on watched object
func (wo watchedObject) getVersion() (string, bool) {
	labels := wo.Obj.GetLabels()
	v, ok := labels[versionLabel]
	return v, ok
}

// getTrack get trace of version from TRACK_ANNOTATION annotation on watched object
func (wo watchedObject) getTrack() string {
	labels := wo.Obj.GetLabels()
	track, ok := labels[trackLabel]
	if !ok {
		return ""
	}
	return track
}

func (wo watchedObject) isIter8AbnRelated() bool {
	labels := wo.Obj.GetLabels()
	iter8, ok := labels[iter8Label]
	if !ok {
		return false
	}
	return strings.ToLower(iter8) == "true"
}
