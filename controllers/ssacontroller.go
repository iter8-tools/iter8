package controllers

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	"k8s.io/client-go/tools/cache"
)

const (
	iter8DetectLabel = "iter8.tools/detect"
	iter8DetectValue = "true"
)

func addSSAEventHandlers(config *Config, client k8sclient.Interface) error {
	for gvkrShort, _ := range config.KnownGVKRs {
		if _, ok := appInformers[gvkrShort]; !ok {
			err := errors.New("appInformers map does not have an informer for type: " + gvkrShort)
			log.Logger.Error(err)
			return err
		}
		appInformers[gvkrShort].Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				s, ok := allSubs.getSubject(obj, gvkrShort)
				if ok {
					s.reconcileSSA(config)
				}
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				s, ok := allSubs.getSubject(newObj, gvkrShort)
				if ok {
					s.reconcileSSA(config)
				}
			},
			DeleteFunc: func(obj interface{}) {
				s, ok := allSubs.getSubject(obj, gvkrShort)
				if ok {
					s.reconcileSSA(config)
				}
			},
		})
	}
	return nil
}
