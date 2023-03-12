package controllers

import (
	"errors"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/tools/cache"
)

const (
	iter8FinalizerStr     = "iter8.tools/finalizer"
	iter8WatchLabel       = "iter8.tools/watch"
	iter8WatchValue       = "true"
	iter8PatchLabel       = "iter8.tools/patch"
	iter8PatchValue       = "true"
	iter8ManagedByLabel   = "app.kubernetes.io/managed-by"
	iter8ManagedByValue   = "iter8"
	iter8KindLabel        = "iter8.tools/kind"
	iter8KindSubjectValue = "subject"
	iter8VersionLabel     = "iter8.tools/version"
	iter8VersionValue     = "v0.14"
)

// informers used to watch application resources,
// one per gvr known to Iter8
var appInformers = make(map[string]informers.GenericInformer)
var subjectInformer corev1.ConfigMapInformer

func initAppResourceInformers(stopCh <-chan struct{}, config *Config, client k8sclient.Interface) error {
	// get defaultResync duration
	defaultResync, err := time.ParseDuration(config.DefaultResync)
	if err != nil {
		e := errors.New("unable to parse defaultResync config option")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return e
	}

	// required labels on application resources that are being watched
	requiredLabels := map[string]string{
		iter8WatchLabel: iter8WatchValue,
	}
	labelSelector := metav1.SetAsLabelSelector(requiredLabels)
	tlo := dynamicinformer.TweakListOptionsFunc(func(opts *metav1.ListOptions) {
		opts.LabelSelector = labelSelector.String()
	})

	// fire up informers
	// config.AppNamespace could equal metav1.NamespaceAll ("") or a specific namespace
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, defaultResync, config.AppNamespace, tlo)
	// factory := dynamicinformer.NewDynamicSharedInformerFactory(client, defaultResync)
	// this map contains an informer for each gvr watched by the controller

	for gvkrShort, gvkr := range config.KnownGVKRs {
		gvkrShort := gvkrShort
		appInformers[gvkrShort] = factory.ForResource(schema.GroupVersionResource{
			Group:    gvkr.Group,
			Version:  gvkr.Version,
			Resource: gvkr.Resource,
		})

		appInformers[gvkrShort].Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				addFinalizer(obj, gvkrShort, client, config)
				s, ok := allSubjects.getSubFromObj(obj, gvkrShort)
				if !ok {
					log.Logger.Warn("no known subject for object. resource short name: ",
						gvkrShort, " object name: ", obj.(*unstructured.Unstructured).GetName())
				}
				s.reconcile(config)
				defer removeFinalizer(obj, gvkrShort, client, config)
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				addFinalizer(newObj, gvkrShort, client, config)
				s, ok := allSubjects.getSubFromObj(newObj, gvkrShort)
				if !ok {
					log.Logger.Warn("no known subject for object. resource short name: ",
						gvkrShort, " object name: ", newObj.(*unstructured.Unstructured).GetName())
				}
				s.reconcile(config)
				defer removeFinalizer(newObj, gvkrShort, client, config)
			},
			DeleteFunc: func(obj interface{}) {
				s, ok := allSubjects.getSubFromObj(obj, gvkrShort)
				if !ok {
					log.Logger.Warn("no known subject for object. resource short name: ",
						gvkrShort, " object name: ", obj.(*unstructured.Unstructured).GetName())
				}
				s.reconcile(config)
				defer removeFinalizer(obj, gvkrShort, client, config)
			},
		})

	}
	log.Logger.Trace("starting app informers factory ...")
	factory.Start(stopCh)
	log.Logger.Trace("started app informers factory ...")
	return nil
}

func initSubjectCMInformer(stopCh <-chan struct{}, config *Config, client k8sclient.Interface) error {
	// get defaultResync duration
	defaultResync, err := time.ParseDuration(config.DefaultResync)
	if err != nil {
		e := errors.New("unable to parse defaultResync config option")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return e
	}

	// specify required labels on configmaps that are being watched
	tlo := internalinterfaces.TweakListOptionsFunc(func(opts *metav1.ListOptions) {
		opts.LabelSelector = labels.Set(map[string]string{
			iter8ManagedByLabel: iter8ManagedByValue,
			iter8KindLabel:      iter8KindSubjectValue,
			iter8VersionLabel:   iter8VersionValue,
		}).String()
	})

	// fire up subject-configmap informer
	// config.AppNamespace could equal metav1.NamespaceAll ("") or a specific namespace
	factory := informers.NewSharedInformerFactoryWithOptions(client, defaultResync, informers.WithNamespace(config.AppNamespace), informers.WithTweakListOptions(tlo))
	si := factory.Core().V1().ConfigMaps()
	si.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			s := allSubjects.makeAndUpdateWith(obj)
			s.reconcile(config)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			s := allSubjects.makeAndUpdateWith(newObj)
			s.reconcile(config)
		},
		DeleteFunc: func(obj interface{}) {
			allSubjects.delete(obj, config)
		},
	})

	log.Logger.Trace("starting app informers factory ...")
	factory.Start(stopCh)
	log.Logger.Trace("started app informers factory ...")
	return nil
}

// Start starts all Iter8 controllers if this pod is the leader
func Start(stopCh <-chan struct{}, client k8sclient.Interface) error {
	config, err := readConfig()
	if err != nil {
		return err
	}

	// validate config
	if err := config.validate(); err != nil {
		return err
	}

	log.Logger.Trace("initing app informers ... ")
	initAppResourceInformers(stopCh, config, client)
	log.Logger.Trace("inited app informers ... ")

	log.Logger.Trace("initing subject informer ... ")
	initSubjectCMInformer(stopCh, config, client)
	log.Logger.Trace("inited subject informer ... ")

	return nil
}
