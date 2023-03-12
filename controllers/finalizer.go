package controllers

import (
	"context"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
)

func addFinalizer(obj interface{}, gvkrShort string, client k8sclient.Interface, config *Config) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		u := obj.(*unstructured.Unstructured)
		// check if there's a deletionTimeStamp
		// if not, add finalizer
		finalizers := append(u.GetFinalizers(), iter8FinalizerStr)
		update := false
		if u.GetDeletionTimestamp() == nil {
			for _, f := range finalizers {
				if f == iter8FinalizerStr {
					break
				}
			}
			update = true
		}

		if update {
			finalizers = append(finalizers, iter8FinalizerStr)
			u.SetFinalizers(finalizers)
			_, e := client.Resource(schema.GroupVersionResource{
				Group:    config.KnownGVKRs[gvkrShort].Group,
				Version:  config.KnownGVKRs[gvkrShort].Version,
				Resource: config.KnownGVKRs[gvkrShort].Resource,
			}).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
			log.Logger.WithStackTrace("failed to add finalizer").Error(e)
			return e
		}

		return nil
	})
	if err != nil {
		log.Logger.WithStackTrace("failed to add finalizer with retry").Error(err)
	}
}

func removeFinalizer(obj interface{}, gvkrShort string, client k8sclient.Interface, config *Config) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		u := obj.(*unstructured.Unstructured)
		finalizers := make([]string, len(u.GetFinalizers()))
		copy(finalizers, u.GetFinalizers())

		for oneMoreLoop := true; oneMoreLoop; {
			// have elements been deleted from finalizers in this loop? No to begin with
			deleted := false
			// go through finalizers and delete iter8 finalizer string if found
			for i, s := range finalizers {
				if s == iter8FinalizerStr {
					slices.Delete(finalizers, i, i+1)
					// deleted something
					deleted = true
					// this loop is over
					break
				}
			}
			// start over if deleted an item in this loop
			// iter8FinalizerStr may be repeated in the slice
			oneMoreLoop = deleted
		}

		if len(finalizers) != len(u.GetFinalizers()) {
			u.SetFinalizers(finalizers)
			_, e := client.Resource(schema.GroupVersionResource{
				Group:    config.KnownGVKRs[gvkrShort].Group,
				Version:  config.KnownGVKRs[gvkrShort].Version,
				Resource: config.KnownGVKRs[gvkrShort].Resource,
			}).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
			log.Logger.WithStackTrace("failed finalizers deletion").Error(e)
			return e
		}
		return nil
	})
	if err != nil {
		log.Logger.WithStackTrace("failed finalizers deletion with retry").Error(err)
	}
}
