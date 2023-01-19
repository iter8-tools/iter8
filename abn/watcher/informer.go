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
func NewIter8Watcher(resourceTypes []schema.GroupVersionResource, namespaces []string) *Iter8Watcher {
	w := &Iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}

	handlerFunc := func(obj interface{}) {
		wo := watchedObject{Obj: obj.(*unstructured.Unstructured)}
		if precond(wo) {
			handle(wo, resourceTypes, w.factories)
		}
	}

	// for each namespace, resource type configure Informer
	for _, ns := range namespaces {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sclient.Client.Dynamic(), 0, ns, nil)
		for _, gvr := range resourceTypes {
			informer := w.factories[ns].ForResource(gvr)
			_, err := informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: handlerFunc,
				UpdateFunc: func(oldObj, obj interface{}) {
					handlerFunc(obj)
				},
				DeleteFunc: handlerFunc,
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
