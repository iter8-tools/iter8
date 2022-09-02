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
	ITER8_ANNOTATION = "iter8.tools/abn"
)

// watchedObject is wrapper for object returned by informer
type watchedObject struct {
	// Obj is the Kubernetes object
	Obj *unstructured.Unstructured
}

// getName gets application name from NAME_LABEL label on watched object
func (wo watchedObject) getName() (string, bool) {
	labels := wo.Obj.GetLabels()
	name, ok := labels[NAME_LABEL]
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
	v, ok := labels[VERSION_LABEL]
	return v, ok
}

// getTrack get trace of version from TRACK_ANNOTATION annotation on watched object
func (wo watchedObject) getTrack() string {
	annotations := wo.Obj.GetAnnotations()
	track, ok := annotations[TRACK_ANNOTATION]
	if !ok {
		return ""
	}
	return track
}

// isReady determines if watched object indicates readiness of version (as indicated by READY_ANNOTATION annotatation)
func (wo watchedObject) isReady() bool {
	annotations := wo.Obj.GetAnnotations()
	ready, ok := annotations[READY_ANNOTATION]
	if !ok {
		return false
	}
	return strings.ToLower(ready) == "true"
}

func (wo watchedObject) isIter8AbnRelated() bool {
	annotations := wo.Obj.GetAnnotations()
	iter8, ok := annotations[ITER8_ANNOTATION]
	if !ok {
		return false
	}
	return strings.ToLower(iter8) == "true"
}
