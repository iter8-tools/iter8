package watcher

import (
	"context"
	"fmt"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func precond(w watchedObject) bool {
	var ok bool

	if !w.isIter8AbnRelated() {
		return false
	}

	_, ok = w.getNamespacedName()
	if !ok {
		return false
	}

	_, ok = w.getVersion()

	return ok
}

func handle(w watchedObject, resourceTypes []schema.GroupVersionResource) {
	application, _ := w.getNamespacedName()
	namespace := w.getNamespace()
	name, _ := w.getName()

	applicationObjs := getApplicationObjects(namespace, name, resourceTypes)
	// there is at least one object (w)

	a, _ := abnapp.Applications.Get(application) // , false)

	abnapp.Applications.Lock(application)
	// clear a.Tracks, a.Versions[*].Track
	// this is necessary because  we keep old versions in memory
	for track := range a.Tracks {
		delete(a.Tracks, track)
	}
	for version := range a.Versions {
		a.Versions[version].Track = nil
	}

	for _, o := range applicationObjs {
		version, _ := o.getVersion()
		v, _ := a.GetVersion(version, true)
		if o.isReady() {
			track := o.getTrack()
			if track != "" {
				v.Track = &track
				if v.Track != nil {
					a.Tracks[*v.Track] = version
				}
			}
		}
	}
	abnapp.Applications.Unlock(application)

	abnapp.Applications.Write(a)
}

func getApplicationObjects(namespace, name string, gvrs []schema.GroupVersionResource) []watchedObject {
	ls := fmt.Sprintf("%s=%s", NAME_LABEL, name)
	watchedObjects := []watchedObject{}
	for _, gvr := range gvrs {
		objs, err := k8sclient.Client.Dynamic().
			Resource(gvr).Namespace(namespace).
			List(context.Background(), metav1.ListOptions{LabelSelector: ls})
		if err != nil {
			log.Logger.Error(err)
			return []watchedObject{}
		}

		wObjs := toWatchedObjectList(objs)
		watchedObjects = append(watchedObjects, wObjs...)
	}
	return watchedObjects
}

func toWatchedObjectList(l *unstructured.UnstructuredList) []watchedObject {
	result := []watchedObject{}
	for _, o := range l.Items {
		w := watchedObject{Obj: &o}
		if precond(w) {
			result = append(result, watchedObject{Obj: &o})
		}
	}
	return result
}
