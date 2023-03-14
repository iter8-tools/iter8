package controllers

import (
	"context"
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
)

func addFinalizer(name string, namespace string, gvrShort string, client k8sclient.Interface, config *Config) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// first, get the object
		u, e := client.Resource(schema.GroupVersionResource{
			Group:    config.KnownGVRs[gvrShort].Group,
			Version:  config.KnownGVRs[gvrShort].Version,
			Resource: config.KnownGVRs[gvrShort].Resource,
		}).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if e != nil {
			return e
		}

		// get old and new finalizers
		oldFinalizers := map[string]bool{}
		newFinalizers := map[string]bool{}
		if u.GetDeletionTimestamp() == nil {
			for _, f := range u.GetFinalizers() {
				oldFinalizers[f] = true
				newFinalizers[f] = true
			}
			// insert Iter8 finalizer
			newFinalizers[iter8FinalizerStr] = true
		}

		// the only way newFinalizers could be of a different length is if
		// oldFinalizers didn't have iter8FinalizerStr
		if len(oldFinalizers) != len(newFinalizers) {
			log.Logger.Trace("oldFinalizers: ", oldFinalizers)
			log.Logger.Trace("newFinalizers: ", newFinalizers)
			finalizers := []string{}
			for key := range newFinalizers {
				finalizers = append(finalizers, key)
			}
			u.SetFinalizers(finalizers)
			log.Logger.Trace("attempting to update resource with finalizer")
			_, e := client.Resource(schema.GroupVersionResource{
				Group:    config.KnownGVRs[gvrShort].Group,
				Version:  config.KnownGVRs[gvrShort].Version,
				Resource: config.KnownGVRs[gvrShort].Resource,
			}).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
			if e != nil {
				log.Logger.WithStackTrace(e.Error()).Error("error while updating resource with finalizer")
			}
			return e
		}

		return nil
	})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			log.Logger.Debug(err)
		} else {
			log.Logger.WithStackTrace(err.Error()).Error(errors.New("failed to add finalizer with retry"))
		}
	}
}

func removeFinalizer(name string, namespace string, gvrShort string, client k8sclient.Interface, config *Config) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// first, get the object
		u, e := client.Resource(schema.GroupVersionResource{
			Group:    config.KnownGVRs[gvrShort].Group,
			Version:  config.KnownGVRs[gvrShort].Version,
			Resource: config.KnownGVRs[gvrShort].Resource,
		}).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if e != nil && kubeerrors.IsNotFound(e) {
			return nil
		} else if e != nil {
			return e
		}

		if u.GetDeletionTimestamp() == nil {
			log.Logger.Trace("object not terminating; will not remove finalizer")
			return nil
		}

		// remove iter8 finalizer
		finalizersMap := map[string]bool{}
		for _, f := range u.GetFinalizers() {
			if f != iter8FinalizerStr {
				finalizersMap[f] = true
			}
		}
		finalizers := make([]string, len(finalizersMap))
		for key := range finalizersMap {
			finalizers = append(finalizers, key)
		}

		// set new finalizers
		u.SetFinalizers(finalizers)
		_, e = client.Resource(schema.GroupVersionResource{
			Group:    config.KnownGVRs[gvrShort].Group,
			Version:  config.KnownGVRs[gvrShort].Version,
			Resource: config.KnownGVRs[gvrShort].Resource,
		}).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})

		if e != nil && kubeerrors.IsNotFound(e) {
			return nil
		}
		return e
	})

	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error(errors.New("failed to delete finalizer with retry"))
	}

}
