package controllers

import (
	"errors"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
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

func initAppResourceInformers(stopCh <-chan struct{}, config *Config, client k8sclient.Interface) error {
	// get defaultResync duration
	defaultResync, err := time.ParseDuration(config.DefaultResync)
	if err != nil {
		e := errors.New("unable to parse defaultResync config option")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return e
	}

	// required labels on application resources that are being watched
	tlo := dynamicinformer.TweakListOptionsFunc(func(opts *metav1.ListOptions) {
		opts.LabelSelector = metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				iter8WatchLabel: iter8WatchValue,
			},
		})
	})

	// fire up informers
	// config.AppNamespace could equal nil or a specific namespace
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, defaultResync, metav1.NamespaceAll, tlo)
	// factory := dynamicinformer.NewDynamicSharedInformerFactory(client, defaultResync)
	// this map contains an informer for each gvr watched by the controller

	for gvrShort, gvr := range config.ResourceTypes {
		gvrShort := gvrShort
		appInformers[gvrShort] = factory.ForResource(schema.GroupVersionResource{
			Group:    gvr.Group,
			Version:  gvr.Version,
			Resource: gvr.Resource,
		})

		appInformers[gvrShort].Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				name := obj.(*unstructured.Unstructured).GetName()
				namespace := obj.(*unstructured.Unstructured).GetNamespace()
				log.Logger.Debug("add called for resource; gvr: ", gvrShort, "; namespace: ", namespace, "; name: ", name)
				addFinalizer(name, namespace, gvrShort, client, config)
				if s := allSubjects.getSubFromObj(obj, gvrShort); s == nil {
					log.Logger.Trace("subject not found; gvr: ",
						gvrShort, "; object name: ", obj.(*unstructured.Unstructured).GetName(),
						"; namespace: ", obj.(*unstructured.Unstructured).GetNamespace())
				} else {
					s.reconcile(config, client)
				}
				defer removeFinalizer(name, namespace, gvrShort, client, config)
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				name := newObj.(*unstructured.Unstructured).GetName()
				namespace := newObj.(*unstructured.Unstructured).GetNamespace()
				log.Logger.Debug("update called for resource; gvr: ", gvrShort, "; namespace: ", namespace, "; name: ", name)
				log.Logger.Debug("finalizers in new obj: ", newObj.(*unstructured.Unstructured).GetFinalizers())
				addFinalizer(name, namespace, gvrShort, client, config)
				if s := allSubjects.getSubFromObj(newObj, gvrShort); s == nil {
					log.Logger.Trace("subject not found; gvr: ",
						gvrShort, "; object name: ", newObj.(*unstructured.Unstructured).GetName(),
						"; namespace: ", newObj.(*unstructured.Unstructured).GetNamespace())
				} else {
					s.reconcile(config, client)
				}
				defer removeFinalizer(name, namespace, gvrShort, client, config)
			},
			DeleteFunc: func(obj interface{}) {
				name := obj.(*unstructured.Unstructured).GetName()
				namespace := obj.(*unstructured.Unstructured).GetNamespace()
				log.Logger.Debug("delete called for resource; gvr: ", gvrShort, "; namespace: ", namespace, "; name: ", name)
				if s := allSubjects.getSubFromObj(obj, gvrShort); s == nil {
					log.Logger.Trace("subject not found; gvr: ",
						gvrShort, "; object name: ", obj.(*unstructured.Unstructured).GetName(),
						"; namespace: ", obj.(*unstructured.Unstructured).GetNamespace())
				} else {
					s.reconcile(config, client)
				}
				defer removeFinalizer(name, namespace, gvrShort, client, config)
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

	// required labels on application resources that are being watched
	tlo := internalinterfaces.TweakListOptionsFunc(func(opts *metav1.ListOptions) {
		opts.LabelSelector = metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				iter8ManagedByLabel: iter8ManagedByValue,
				iter8KindLabel:      iter8KindSubjectValue,
				iter8VersionLabel:   iter8VersionValue,
			},
		})
	})

	// fire up subject-configmap informer
	// config.AppNamespace could equal metav1.NamespaceAll ("") or a specific namespace
	factory := informers.NewSharedInformerFactoryWithOptions(client, defaultResync, informers.WithNamespace(metav1.NamespaceAll), informers.WithTweakListOptions(tlo))
	si := factory.Core().V1().ConfigMaps()
	si.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Logger.Trace("in subject add func")
			log.Logger.Trace("making and updating subject")
			s := allSubjects.makeAndUpdateWith(obj.(*corev1.ConfigMap))
			if s == nil {
				log.Logger.Error("unable to create subject from configmap; ", "namespace: ", obj.(*corev1.ConfigMap).Namespace, "; name: ", obj.(*corev1.ConfigMap).Name)
				return
			}
			s.reconcile(config, client)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			s := allSubjects.makeAndUpdateWith(newObj.(*corev1.ConfigMap))
			if s == nil {
				log.Logger.Error("unable to create subject from configmap; ", "namespace: ", newObj.(*corev1.ConfigMap).Namespace, "; name: ", newObj.(*corev1.ConfigMap).Name)
				return
			}
			s.reconcile(config, client)
		},
		DeleteFunc: func(obj interface{}) {
			allSubjects.delete(obj.(*corev1.ConfigMap), config, client)
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

	// get leaderStatus
	if leaderStatus, err := leaderIsMe(); err != nil {
		return err
	} else {
		log.Logger.Info("leader: ", leaderStatus)
	}

	log.Logger.Trace("initing app informers ... ")
	initAppResourceInformers(stopCh, config, client)
	log.Logger.Trace("inited app informers ... ")

	log.Logger.Trace("initing subject informer ... ")
	initSubjectCMInformer(stopCh, config, client)
	log.Logger.Trace("inited subject informer ... ")

	return nil
}
