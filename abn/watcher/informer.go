package watcher

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// MultiInformer holds all of the informers for all namespaces
type MultiInformer struct {
	informersByKey             map[string]informers.GenericInformer
	informerFactroyByNamespace map[string]dynamicinformer.DynamicSharedInformerFactory
}

// addInformer adds an informer
func (informer *MultiInformer) addInformer(ns string, gvr schema.GroupVersionResource) {
	informer.informersByKey[key(ns, gvr)] = informer.informerFactroyByNamespace[ns].ForResource(gvr)
}

// key computes a unique key for an informer from the gvr and namespace
func key(ns string, gvr schema.GroupVersionResource) string {
	return ns + "/" + gvr.Group + "." + gvr.Version + "." + gvr.Resource
}

// NewInformer creates a new informer watching the desired resoures in each of the desired namespaces
func NewInformer(kd *driver.KubeDriver, types []schema.GroupVersionResource, namespaces []string) *MultiInformer {
	informer := &MultiInformer{
		informersByKey:             make(map[string]informers.GenericInformer, len(types)*len(namespaces)),
		informerFactroyByNamespace: make(map[string]dynamicinformer.DynamicSharedInformerFactory, len(namespaces)),
	}
	for _, ns := range namespaces {
		informer.informerFactroyByNamespace[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(kd.DynamicClient, 0, ns, nil)
		for _, gvr := range types {
			log.Logger.Debugf("Configured watcher for <%s> in namespace %s", gvr.String(), ns)
			informer.addInformer(ns, gvr)
		}
	}
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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

	return informer
}

// AddEventHandler adds event handler to each informer
func (informer *MultiInformer) AddEventHandler(handler cache.ResourceEventHandlerFuncs) {
	for _, i := range informer.informersByKey {
		i.Informer().AddEventHandler(handler)
	}
}

// Start calls Start on each informer
func (informer *MultiInformer) Start(stopCh <-chan struct{}) {
	for _, i := range informer.informerFactroyByNamespace {
		i.Start(stopCh)
	}
}