package controllers

import (
	"errors"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
)

// all subjects creaetd by user in every namespace managed by Iter8
var allSubs = make(allSubjects)

// informers used to detect application resources,
// one per gvr known to Iter8
var appInformers = make(map[string]informers.GenericInformer)

func initAppInformers(stopCh <-chan struct{}, config *Config, client k8sclient.Interface) error {
	// get defaultResync duration
	defaultResync, err := time.ParseDuration(config.DefaultResync)
	if err != nil {
		e := errors.New("unable to parse defaultResync config option")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return e
	}

	// required labels on application resources that are being watched
	requiredLabels := map[string]string{
		iter8DetectLabel: iter8DetectValue,
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
		appInformers[gvkrShort] = factory.ForResource(schema.GroupVersionResource{
			Group:    gvkr.Group,
			Version:  gvkr.Version,
			Resource: gvkr.Resource,
		})
	}
	log.Logger.Trace("starting app informers factory ...")
	factory.Start(stopCh)
	log.Logger.Trace("started app informers factory ...")
	return nil
}

// Start starts all Iter8 controllers if this pod is the leader
func Start(stopCh chan struct{}, client k8sclient.Interface) error {
	config, err := readConfig()
	if err != nil {
		return err
	}

	// validate config
	if err := config.validate(); err != nil {
		return err
	}

	// everyone could start other controllers, like the abn controller
	// for e.g.,
	// go startABnController(stopCh, config, client)

	log.Logger.Trace("starting app informers ... ")
	initAppInformers(stopCh, config, client)
	log.Logger.Trace("started app informers ... ")

	log.Logger.Trace("spawning subject cm controller ... ")
	go startSubjectCMController(stopCh, config, client)
	log.Logger.Trace("spawned subject cm controller ... ")

	log.Logger.Trace("checking if leader is me ...")
	// Only the leader pod starts SSA controller
	if leaderIsMe() {
		log.Logger.Trace("leader is me ... ")

		log.Logger.Trace("invoking add SSA event handlers ... ")
		// add server-side apply event handlers
		return addSSAEventHandlers(config, client)

	}

	return nil
}
