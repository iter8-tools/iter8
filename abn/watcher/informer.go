package watcher

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"fmt"

	"github.com/iter8-tools/iter8/abn/k8sclient"
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
// func NewIter8Watcher(resourceTypes []schema.GroupVersionResource, namespaces []string) *Iter8Watcher {
func NewIter8Watcher(configFile string) *Iter8Watcher {
	c := readServiceConfig(configFile)

	w := &Iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}

	handlerFunc := func(action string, validNames []string) func(obj interface{}) {
		return func(obj interface{}) {
			o := obj.(*unstructured.Unstructured)
			if containsString(validNames, o.GetName()) {
				handle(action, o, c, w.factories)
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
			validNames := validObjectNames(nm, details.MaxNumCandidates)
			for _, r := range details.Resources {
				byResource[r.GroupVersionResource] = append(byResource[r.GroupVersionResource], validNames...)
			}
		}

		// create informer for each resource in namespace
		for r, validNames := range byResource {
			informer := w.factories[ns].ForResource(r)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: handlerFunc("ADD", validNames),
				UpdateFunc: func(oldObj, obj interface{}) {
					handlerFunc("UPDATE", validNames)(obj)
				},
				DeleteFunc: handlerFunc("DELETE", validNames),
			})
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
func validObjectNames(application string, maxNum int) []string {
	expectedObjectNames := make([]string, maxNum+1)
	expectedObjectNames[0] = application
	for i := 1; i <= maxNum; i++ {
		expectedObjectNames[i] = fmt.Sprintf("%s-candidate-%d", application, i)
	}
	return expectedObjectNames
}

// trackNames creates list of expected trackNames
// based on the assumption that the track names are the same as the object names
func trackNames(application string, appConfig appDetails) []string {
	return validObjectNames(application, appConfig.MaxNumCandidates)
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
