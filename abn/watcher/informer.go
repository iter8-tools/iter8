package watcher

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"fmt"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// Iter8Watcher enables creation of informers needed by the abn service
type Iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

// NewIter8Watcher returns a watcher for iter8 related objects
func NewIter8Watcher(configFile string) *Iter8Watcher {
	c := readServiceConfig(configFile)

	w := &Iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}

	addHandlerFunc := func(validNames []string, gvr schema.GroupVersionResource) func(obj interface{}) {
		return func(obj interface{}) {
			o := obj.(*unstructured.Unstructured)
			log.Logger.Tracef("add handler called for %s/%s (%s)", o.GetNamespace(), o.GetName(), o.GetKind())
			defer log.Logger.Tracef("add handler completed for %s/%s (%s)", o.GetNamespace(), o.GetName(), o.GetKind())

			if containsString(validNames, o.GetName()) {
				handle(o, c, w.factories, gvr)
			}
		}
	}

	updateHandlerFunc := func(validNames []string, gvr schema.GroupVersionResource) func(oldObj, obj interface{}) {
		return func(oldObj, obj interface{}) {
			o := obj.(*unstructured.Unstructured)
			log.Logger.Tracef("update handler called for %s/%s (%s)", o.GetNamespace(), o.GetName(), o.GetKind())
			defer log.Logger.Tracef("update handler completed for %s/%s (%s)", o.GetNamespace(), o.GetName(), o.GetKind())

			if containsString(validNames, o.GetName()) {
				handle(o, c, w.factories, gvr)
			}
		}
	}

	deleteHandlerFunc := func(validNames []string, gvr schema.GroupVersionResource) func(obj interface{}) {
		return func(obj interface{}) {
			o := obj.(*unstructured.Unstructured)
			log.Logger.Tracef("delete handler called for %s/%s (%s)", o.GetNamespace(), o.GetName(), o.GetKind())
			defer log.Logger.Tracef("delete handler completed for %s/%s (%s)", o.GetNamespace(), o.GetName(), o.GetKind())

			if containsString(validNames, o.GetName()) {
				handle(o, c, w.factories, gvr)
			}
		}
	}

	// for each namespace:
	//   identify all resources to watch (create informer for each), and
	//   names of all resources to watch
	//   create informerfactory
	for ns, apps := range c {
		var byResource map[schema.GroupVersionResource]([]string) = map[schema.GroupVersionResource]([]string){}
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sclient.Client.Dynamic(), 0, ns, nil)

		// identify resources expected in namespace and list of resource names expected
		// the resources will be used to create informers
		// the list of names will be used to filter the objects that trigger the informer
		for nm, details := range apps {
			validNames := getValidObjectNames(nm, details.MaxNumCandidates)
			for _, r := range details.Resources {
				byResource[r.GroupVersionResource] = append(byResource[r.GroupVersionResource], validNames...)
			}
		}

		// create informer for each resource in namespace
		for gvr, validNames := range byResource {
			informer := w.factories[ns].ForResource(gvr)
			_, err := informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    addHandlerFunc(validNames, gvr),
				UpdateFunc: updateHandlerFunc(validNames, gvr),
				DeleteFunc: deleteHandlerFunc(validNames, gvr),
			})

			if err != nil {
				log.Logger.Error(fmt.Sprintf("cannot add event handler for namespace \"%s\" and GVR \"%s\": \"%s\"", ns, gvr, err))
			}
		}
	}
	return w
}

// Start starts the watcher
func (watcher *Iter8Watcher) Start(stopChannel chan struct{}) {
	for _, f := range watcher.factories {
		f.Start(stopChannel)
	}
}

// valid object names for an application
// based on the assumption that the names are of the form:
//
//	app, app-candidate-i for i = 1, 2, ..., maxNumCandidates
func getValidObjectNames(application string, maxNum int) []string {
	expectedObjectNames := make([]string, maxNum+1)
	expectedObjectNames[0] = application
	for i := 1; i <= maxNum; i++ {
		expectedObjectNames[i] = fmt.Sprintf("%s-candidate-%d", application, i)
	}
	return expectedObjectNames
}

// getTrackNames creates list of expected getTrackNames
// based on the assumption that the track names are the same as the object names
func getTrackNames(application string, appConfig appDetails) []string {
	return getValidObjectNames(application, appConfig.MaxNumCandidates)
}

// containsString determines if an array contains a specific string
func containsString(array []string, s string) bool {
	for _, v := range array {
		if v == s {
			return true
		}
	}
	return false
}
