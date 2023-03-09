package controllers

import (
	"errors"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/tools/cache"
)

const (
	iter8ManagedByLabel = "app.kubernetes.io/managed-by"
	iter8ManagedByValue = "iter8"
	iter8KindLabel      = "iter8.tools/kind"
	iter8KindValue      = "subject"
	iter8VersionLabel   = "iter8.tools/version"
	iter8VersionValue   = "v0.14"
)

func startSubjectCMController(stopCh chan struct{}, config *Config, client k8sclient.Interface) error {
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
			iter8KindLabel:      iter8KindValue,
			iter8VersionLabel:   iter8VersionValue,
		}).String()
	})

	// fire up subject-configmap informer
	// config.AppNamespace could equal metav1.NamespaceAll ("") or a specific namespace
	factory := informers.NewSharedInformerFactoryWithOptions(client, defaultResync, informers.WithNamespace(config.AppNamespace), informers.WithTweakListOptions(tlo))
	// factory := informers.NewSharedInformerFactory(client, defaultResync)
	inf := factory.Core().V1().ConfigMaps().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			allSubs.makeSubject(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			allSubs.makeSubject(new)
		},
		DeleteFunc: func(obj interface{}) {
			allSubs.deleteSubject(obj)
		},
	})

	log.Logger.Trace("starting subject cm informer factory ...")
	factory.Start(stopCh)
	log.Logger.Trace("started subject cm informer factory ...")
	return nil
}
