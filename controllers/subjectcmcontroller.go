package controllers

import (
	"errors"
	"time"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// required labels on configmaps that are being watched
	requiredLabels := map[string]string{
		iter8ManagedByLabel: iter8ManagedByValue,
		iter8KindLabel:      iter8KindValue,
		iter8VersionLabel:   iter8VersionValue,
	}
	labelSelector := metav1.SetAsLabelSelector(requiredLabels)
	tlo := internalinterfaces.TweakListOptionsFunc(func(opts *metav1.ListOptions) {
		opts.LabelSelector = labelSelector.String()
	})

	// fire up subject-configmap informer
	// config.AppNamespace could equal metav1.NamespaceAll ("") or a specific namespace
	factory := informers.NewSharedInformerFactoryWithOptions(client, defaultResync, informers.WithNamespace(config.AppNamespace), informers.WithTweakListOptions(tlo))
	inf := factory.Core().V1().ConfigMaps().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(cmObj interface{}) {
			allSubs.makeSubject(cmObj)
		},
		UpdateFunc: func(oldCMObj interface{}, newCMObj interface{}) {
			allSubs.makeSubject(newCMObj)
		},
		DeleteFunc: func(cmObj interface{}) {
			allSubs.deleteSubject(cmObj)
		},
	})
	factory.Start(stopCh)
	return nil
}
