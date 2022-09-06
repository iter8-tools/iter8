package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher(resourceTypes []schema.GroupVersionResource, namespaces []string, groupConfig chartGroupConfig) *iter8Watcher {
	w := &iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}
	// for each namespace, resource type configure Informer
	for _, ns := range namespaces {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sclient.dynamicClient, 0, ns, nil)
		for _, gvr := range resourceTypes {
			informer := w.factories[ns].ForResource(gvr)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					fmt.Println("Add:", obj)
					// Add(WatchedObject{
					// 	Obj:    obj.(*unstructured.Unstructured),
					// 	Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					// })
				},
				UpdateFunc: func(oldObj, obj interface{}) {
					fmt.Println("Update:", obj)
					// Update(WatchedObject{
					// 	Obj:    obj.(*unstructured.Unstructured),
					// 	Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					// })
				},
				DeleteFunc: func(obj interface{}) {
					fmt.Println("Delete:", obj)
					// Delete(WatchedObject{
					// 	Obj:    obj.(*unstructured.Unstructured),
					// 	Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					// })
				},
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
