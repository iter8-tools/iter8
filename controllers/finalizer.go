package controllers

import (
	"context"
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
)

const (
	// for application resources
	iter8FinalizerStr = "iter8.tools/finalizer"
)

// add Iter8 finalizer to an application resource
func addFinalizer(name string, namespace string, gvrShort string, client k8sclient.Interface, config *Config) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// first, get the object
		u, e := client.Resource(schema.GroupVersionResource{
			Group:    config.ResourceTypes[gvrShort].Group,
			Version:  config.ResourceTypes[gvrShort].Version,
			Resource: config.ResourceTypes[gvrShort].Resource,
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
				Group:    config.ResourceTypes[gvrShort].Group,
				Version:  config.ResourceTypes[gvrShort].Version,
				Resource: config.ResourceTypes[gvrShort].Resource,
			}).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
			if e != nil {
				log.Logger.WithStackTrace(e.Error()).Error("error while updating resource with finalizer")
			} else {
				// broadcast event

				// get resource for event broadcasting
				r, err := client.Resource(schema.GroupVersionResource{
					Group:    config.ResourceTypes[gvrShort].Group,
					Version:  config.ResourceTypes[gvrShort].Version,
					Resource: config.ResourceTypes[gvrShort].Resource,
				}).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err != nil {
					log.Logger.Warnf("could not get pod with name %s in namespace %s", name, namespace)
				} else {
					broadcastEvent(r, corev1.EventTypeNormal, "Added Iter8 finalizer", "Added Iter8 finalizer for resource", client)
				}
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

			// get resource for event broadcasting
			r, err := client.Resource(schema.GroupVersionResource{
				Group:    config.ResourceTypes[gvrShort].Group,
				Version:  config.ResourceTypes[gvrShort].Version,
				Resource: config.ResourceTypes[gvrShort].Resource,
			}).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				log.Logger.Warnf("could not get pod with name %s in namespace %s", name, namespace)
			}

			broadcastEvent(r, corev1.EventTypeWarning, "Failed to add Iter8 finalizer", "Failed to add Iter8 finalizer for resource", client)
		}
	}
}

// remove Iter8 finalizer from an application resource
func removeFinalizer(name string, namespace string, gvrShort string, client k8sclient.Interface, config *Config) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// first, get the object
		u, e := client.Resource(schema.GroupVersionResource{
			Group:    config.ResourceTypes[gvrShort].Group,
			Version:  config.ResourceTypes[gvrShort].Version,
			Resource: config.ResourceTypes[gvrShort].Resource,
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

		// remove iter8 finalizer if present
		var finalizers []string
		for _, f := range u.GetFinalizers() {
			if f != iter8FinalizerStr {
				finalizers = append(finalizers, f)
			}
		}

		if len(finalizers) < len(u.GetFinalizers()) {
			// update finalizers in the object
			// we do not want to remove non-Iter8 finalizers
			if len(finalizers) == 0 {
				u.SetFinalizers(nil)
			} else {
				u.SetFinalizers(finalizers)
			}

			// update object
			_, e = client.Resource(schema.GroupVersionResource{
				Group:    config.ResourceTypes[gvrShort].Group,
				Version:  config.ResourceTypes[gvrShort].Version,
				Resource: config.ResourceTypes[gvrShort].Resource,
			}).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})

			if e != nil {
				// if object has been deleted, return
				if kubeerrors.IsNotFound(e) {
					return nil
				}
			}
		}

		return e
	})

	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error(errors.New("failed to delete Iter8 finalizer"))

		// get resource for event broadcasting
		r, err := client.Resource(schema.GroupVersionResource{
			Group:    config.ResourceTypes[gvrShort].Group,
			Version:  config.ResourceTypes[gvrShort].Version,
			Resource: config.ResourceTypes[gvrShort].Resource,
		}).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			log.Logger.Warnf("could not get resource with name %s in namespace %s", name, namespace)
		}

		broadcastEvent(r, corev1.EventTypeWarning, "Failed to delete Iter8 finalizer", "Failed to delete Iter8 finalizer for resource", client)
	}
}
