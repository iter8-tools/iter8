// Package controllers provides Iter8 controller for reconciling Iter8 routemap resources
package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	"github.com/iter8-tools/iter8/controllers/storageclient/badgerdb"
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
	// for application resources
	iter8WatchLabel = "iter8.tools/watch"
	iter8WatchValue = "true"

	// for routemap resource
	iter8ManagedByLabel    = "app.kubernetes.io/managed-by"
	iter8ManagedByValue    = "iter8"
	iter8KindLabel         = "iter8.tools/kind"
	iter8KindRoutemapValue = "routemap"
	iter8VersionLabel      = "iter8.tools/version"

	// for persistent volume
	metricsPath = "/metrics"
)

// informers used to watch application resources,
// one per resource type
var appInformers = make(map[string]informers.GenericInformer)

// initAppResourceInformers initializes app resource informers
func initAppResourceInformers(stopCh <-chan struct{}, config *Config, client k8sclient.Interface) error {
	// get defaultResync duration
	defaultResync, err := time.ParseDuration(config.DefaultResync)
	if err != nil {
		e := errors.New("unable to parse defaultResync config option")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return e
	}

	// required labels on application resources that will be watched
	tlo := dynamicinformer.TweakListOptionsFunc(func(opts *metav1.ListOptions) {
		opts.LabelSelector = metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				iter8WatchLabel: iter8WatchValue,
			},
		})
	})

	// factory is used to create informers
	// factory will be cluster-scoped or namespace-scoped as specified in config
	var factory dynamicinformer.DynamicSharedInformerFactory
	var ns string
	if config.ClusterScoped {
		ns = metav1.NamespaceAll
	} else {
		// namespace scoped
		var ok bool
		if ns, ok = os.LookupEnv(podNamespaceEnvVariable); !ok {
			return errors.New("unable to get pod namespace")
		}
	}
	factory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, defaultResync, ns, tlo)

	// handle is an idempotent handler function that is used for any app resource related event
	// 1. deal with app resource finalizers
	// 2. reconcile routemap corresponding to this resource
	handle := func(obj interface{}, gvrShort string, event string) {
		name := obj.(*unstructured.Unstructured).GetName()
		namespace := obj.(*unstructured.Unstructured).GetNamespace()
		log.Logger.Debug(event+" occurred for resource; gvr: ", gvrShort, "; namespace: ", namespace, "; name: ", name)
		addFinalizer(name, namespace, gvrShort, client, config)
		defer removeFinalizer(name, namespace, gvrShort, client, config)
		if s := allRoutemaps.getRoutemapFromObj(obj, gvrShort); s == nil {
			log.Logger.Trace("routemap not found; gvr: ",
				gvrShort, "; object name: ", obj.(*unstructured.Unstructured).GetName(),
				"; namespace: ", obj.(*unstructured.Unstructured).GetNamespace())
		} else {
			s.reconcile(config, client)
		}
	}

	// create an informer per gvr specified in Iter8 config
	for gvrShort, gvr := range config.ResourceTypes {
		gvrShort := gvrShort
		appInformers[gvrShort] = factory.ForResource(schema.GroupVersionResource{
			Group:    gvr.Group,
			Version:  gvr.Version,
			Resource: gvr.Resource,
		})

		// add event handler for the newly minted informer
		if _, err = appInformers[gvrShort].Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				handle(obj, gvrShort, "add")
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				handle(newObj, gvrShort, "update")
			},
			DeleteFunc: func(obj interface{}) {
				handle(obj, gvrShort, "delete")
			},
		}); err != nil {
			e := errors.New("unable to create event handlers for app informers")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
	}

	log.Logger.Trace("starting app informers factory...")
	// start all informers
	factory.Start(stopCh)
	log.Logger.Trace("started app informers factory...")

	return nil
}

// initRoutemapCMInformer initializes routemap informers
func initRoutemapCMInformer(stopCh <-chan struct{}, config *Config, client k8sclient.Interface) error {
	// get defaultResync duration
	defaultResync, err := time.ParseDuration(config.DefaultResync)
	if err != nil {
		e := errors.New("unable to parse defaultResync config option")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return e
	}

	// required labels on routemaps that will be watched
	tlo := internalinterfaces.TweakListOptionsFunc(func(opts *metav1.ListOptions) {
		opts.LabelSelector = metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				iter8ManagedByLabel: iter8ManagedByValue,
				iter8KindLabel:      iter8KindRoutemapValue,
				iter8VersionLabel:   base.MajorMinor,
			},
		})
	})

	// factory is used to create routemap informer
	// factory can be cluster-scoped or namespace-scoped
	var factory informers.SharedInformerFactory
	var ns string
	if config.ClusterScoped {
		ns = metav1.NamespaceAll
	} else {
		// namespace scoped
		var ok bool
		if ns, ok = os.LookupEnv(podNamespaceEnvVariable); !ok {
			return errors.New("unable to get pod namespace")
		}
	}
	factory = informers.NewSharedInformerFactoryWithOptions(client, defaultResync, informers.WithNamespace(ns), informers.WithTweakListOptions(tlo))

	// handle is used during creation and update of routemaps
	// 1. make and update the routemap in allRoutemaps
	// 2. reconcile routemap
	// unlike app resource handle func, routemap handle func is not used for delete events
	handle := func(obj interface{}, event string) {
		log.Logger.Trace(event + " event for routemap")
		s := allRoutemaps.makeAndUpdateWith(obj.(*corev1.ConfigMap), config)
		if s == nil {
			log.Logger.Error("unable to create routemap from configmap; ", "namespace: ", obj.(*corev1.ConfigMap).Namespace, "; name: ", obj.(*corev1.ConfigMap).Name)
			return
		}
		s.reconcile(config, client)
	}

	// create a routemap informer and add handler func
	si := factory.Core().V1().ConfigMaps()
	if _, err = si.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handle(obj, "add")
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			handle(newObj, "update")
		},
		DeleteFunc: func(obj interface{}) {
			allRoutemaps.delete(obj.(*corev1.ConfigMap))
		},
	}); err != nil {
		e := errors.New("unable to create event handlers for routemap informer")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	log.Logger.Trace("starting routemap informer factory...")
	factory.Start(stopCh)
	log.Logger.Trace("started routemap informer factory...")
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

	// get pod for event broadcasting
	ns, ok := os.LookupEnv(podNamespaceEnvVariable)
	if !ok {
		log.Logger.Warnf("could not get pod namespace from environment variable %s", podNamespaceEnvVariable)
	}
	name, ok := os.LookupEnv(podNameEnvVariable)
	if !ok {
		log.Logger.Warnf("could not get pod name from environment variable %s", podNameEnvVariable)
	}
	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Logger.Warnf("could not get pod with name %s in namespace %s", name, ns)
	}

	if config.Persist {
		// TODO: expose badgerDB options in config?
		dbClient, err := badgerdb.GetClient(badger.DefaultOptions(metricsPath))

		fmt.Println(dbClient, err)
	}

	log.Logger.Trace("initing app informers... ")
	if err = initAppResourceInformers(stopCh, config, client); err != nil {
		broadcastEvent(pod, corev1.EventTypeWarning, "Failed to start Iter8 app informers", "Failed to start Iter8 app informers", client)

		return err
	}
	log.Logger.Trace("inited app informers... ")

	log.Logger.Trace("initing routemap informer... ")
	if err = initRoutemapCMInformer(stopCh, config, client); err != nil {
		broadcastEvent(pod, corev1.EventTypeWarning, "Failed to start Iter8 routemap informers", "Failed to start Iter8 routemap informers", client)

		return err
	}
	log.Logger.Trace("inited routemap informer... ")

	return nil
}
