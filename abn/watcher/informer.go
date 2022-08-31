package watcher

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"github.com/iter8-tools/iter8/driver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type Iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func NewIter8Watcher(kd *driver.KubeDriver, resourceTypes []schema.GroupVersionResource, namespaces []string) *Iter8Watcher {
	w := &Iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}
	// for each namespace, resource type configure Informer
	for _, ns := range namespaces {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(kd.DynamicClient, 0, ns, nil)
		for _, gvr := range resourceTypes {
			informer := w.factories[ns].ForResource(gvr)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					addObject(WatchedObject{Obj: obj.(*unstructured.Unstructured)})
				},
				UpdateFunc: func(oldObj, obj interface{}) {
					updateObject(WatchedObject{Obj: obj.(*unstructured.Unstructured)})
				},
				DeleteFunc: func(obj interface{}) {
					deleteObject(WatchedObject{Obj: obj.(*unstructured.Unstructured)})
				},
			})
		}
	}
	return w
}

func (watcher *Iter8Watcher) Start(stopChannel chan struct{}) {
	for _, f := range watcher.factories {
		f.Start(stopChannel)
	}
}
