package watcher

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"github.com/iter8-tools/iter8/abn/application"
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
					Add(WatchedObject{
						Obj:    obj.(*unstructured.Unstructured),
						Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					})
				},
				UpdateFunc: func(oldObj, obj interface{}) {
					Update(WatchedObject{
						Obj:    obj.(*unstructured.Unstructured),
						Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					})
				},
				DeleteFunc: func(obj interface{}) {
					Delete(WatchedObject{
						Obj:    obj.(*unstructured.Unstructured),
						Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					})
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
