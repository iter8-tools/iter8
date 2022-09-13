package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

var addObjectInvocations int = 0
var updateObjectInvocations int = 0
var deleteObjectInvocations int = 0

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher(resourceTypes []schema.GroupVersionResource, namespaces []string, groupConfig chartGroupConfig) *iter8Watcher {
	w := &iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}
	// for each namespace, resource type configure Informer
	for _, ns := range namespaces {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(Client.dynamicClient, 0, ns, nil)
		for _, gvr := range resourceTypes {
			informer := w.factories[ns].ForResource(gvr)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					addObject(obj)
					// Add(WatchedObject{
					// 	Obj:    obj.(*unstructured.Unstructured),
					// 	Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					// })
				},
				UpdateFunc: func(oldObj, obj interface{}) {
					updateObject(oldObj, obj)
					// Update(WatchedObject{
					// 	Obj:    obj.(*unstructured.Unstructured),
					// 	Writer: &application.ApplicationReaderWriter{Client: kd.Clientset},
					// })
				},
				DeleteFunc: func(obj interface{}) {
					deleteObject(obj)
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

func addObject(obj interface{}) {
	log.Logger.Debug("Add:", obj)
	addObjectInvocations++
}

func updateObject(oldObj, obj interface{}) {
	log.Logger.Debug("Update:", obj)
	updateObjectInvocations++
}

func deleteObject(obj interface{}) {
	log.Logger.Debug("Delete:", obj)
	deleteObjectInvocations++
}

func (watcher *iter8Watcher) start(stopChannel chan struct{}) {
	for _, f := range watcher.factories {
		f.Start(stopChannel)
	}
}
