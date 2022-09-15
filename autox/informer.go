package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

var addObject = func(obj interface{}) {
	log.Logger.Debug("Add:", obj)
}

var updateObject = func(oldObj, obj interface{}) {
	log.Logger.Debug("Update:", oldObj, obj)
}

var deleteObject = func(obj interface{}) {
	log.Logger.Debug("Delete:", obj)
}

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher(resourceTypes []schema.GroupVersionResource, namespaces []string, groupConfig chartGroupConfig) *iter8Watcher {
	w := &iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}
	// for each namespace, resource type configure Informer
	for _, ns := range namespaces {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sClient.dynamicClient, 0, ns, nil)
		for _, gvr := range resourceTypes {
			informer := w.factories[ns].ForResource(gvr)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    addObject,
				UpdateFunc: updateObject,
				DeleteFunc: deleteObject,
			})
		}
	}
	return w
}

func (watcher *iter8Watcher) start(stopChannel chan struct{}) {
	for _, f := range watcher.factories {
		f.Start(stopChannel)
	}
}
