package watcher

import (
	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

// precond is set of preconditions that must hold true before an object is considered.
// It must have:
//   - label 'iter8.tools/abn' set to true indicating the resource should be inspected further
//   - label 'app.kubernetes.io/name' identifying the name of the application to which the resource belongs
//   - label 'app.kubernetes.io/version' identifying the name of the version  to which the resource belongs
func precond(w watchedObject) bool {
	var ok bool

	if !w.isIter8AbnRelated() {
		return false
	}

	application, ok := w.getNamespacedName()
	if !ok || application == "" {
		return false
	}

	version, ok := w.getVersion()
	if !ok || version == "" {
		return false
	}

	return true
}

// handle constructs the application object from the objects currently in the cluster
func handle(w watchedObject, resourceTypes []schema.GroupVersionResource, informerFactories map[string]dynamicinformer.DynamicSharedInformerFactory) {
	application, _ := w.getNamespacedName()
	namespace := w.getNamespace()
	name, _ := w.getName()
	version, _ := w.getVersion()
	log.Logger.Tracef("handle called for %s (%s)", application, version)

	applicationObjs := getApplicationObjects(namespace, name, resourceTypes, informerFactories)
	// there is at least one object (w)

	a, _ := abnapp.Applications.Read(application)

	abnapp.Applications.Lock(application)
	defer abnapp.Applications.Unlock(application)
	// clear a.Tracks, a.Versions[*].Track
	// this is necessary because  we keep old versions in memory
	a.ClearTracks()

	for _, o := range applicationObjs {
		version, _ := o.getVersion()
		a.GetVersion(version, true) // make sure version object created
		track := o.getTrack()
		if track != "" {
			a.Tracks[track] = version
		}
	}
}

// getApplicationObjects gets all the objects related to the application based on label app.kubernetes.io/name
func getApplicationObjects(namespace, name string, gvrs []schema.GroupVersionResource, informerFactories map[string]dynamicinformer.DynamicSharedInformerFactory) []watchedObject {
	// define selector
	selector := labels.NewSelector()
	reqSpec := []struct {
		key  string
		op   selection.Operator
		vals []string
	}{
		{key: iter8Label, op: selection.Equals, vals: []string{"true"}},
		{key: nameLabel, op: selection.Equals, vals: []string{name}},
		{key: versionLabel, op: selection.Exists, vals: []string{}},
	}
	for _, rs := range reqSpec {
		req, err := labels.NewRequirement(rs.key, rs.op, rs.vals)
		if err != nil {
			log.Logger.Warn(err)
			return []watchedObject{}
		}
		selector = selector.Add(*req)
	}

	watchedObjects := []watchedObject{}
	for _, gvr := range gvrs {
		lister := informerFactories[namespace].ForResource(gvr).Lister()
		objs, err := lister.List(selector)
		if err != nil {
			log.Logger.Warn(err)
			continue
		}
		for _, obj := range objs {
			wo := watchedObject{Obj: obj.(*unstructured.Unstructured)}
			if precond(wo) {
				watchedObjects = append(watchedObjects, wo)
			}
		}
	}
	return watchedObjects
}
