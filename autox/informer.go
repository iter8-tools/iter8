package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"fmt"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

const (
	autoxLabel = "iter8.tools/autox-group"
)

func getReleaseName(chartGroupName string, chartName string) string {
	return fmt.Sprintf("autox-%s-%s", chartGroupName, chartName)
}

func releaseHelmChart(releaseName string, chart chart) {
	// TODO: check if there is a preexisting Helm release

	// TODO: release Helm chart

	log.Logger.Debug("Release chart:", releaseName)
}

func deleteHelmChart(releaseName string, chart chart) {
	// TODO: check if there is a preexisting Helm release

	// TODO: delete Helm chart

	log.Logger.Debug("Delete chart:", releaseName)
}

func iterateCharts(appName string, cgc chartGroupConfig, callback func(releaseName string, chart chart)) {
	if cg, ok := cgc[appName]; ok {
		for chartName, chart := range cg.Charts {
			releaseName := getReleaseName(appName, chartName)

			callback(releaseName, chart)
		}
	} else {
		// TODO: what log level should this be?
		log.Logger.Debug("AutoX should make a Helm release for app \"", appName, "\" but no Helm charts were provided in the chartGroupConfig")
	}
}

var addObject = func(cgc chartGroupConfig) func(obj interface{}) {
	return func(obj interface{}) {
		log.Logger.Debug("Add:", obj)

		uObj := obj.(*unstructured.Unstructured)
		appName := uObj.GetName()
		// example label: iter8.tools/autox-group=hello
		labels := uObj.GetLabels()

		// check if the app name matches the name in the autox label
		if autoxLabelName, ok := labels[autoxLabel]; ok && appName == autoxLabelName {
			// Release Helm charts
			iterateCharts(appName, cgc, releaseHelmChart)
		}
	}
}

var updateObject = func(cgc chartGroupConfig) func(oldObj, obj interface{}) {
	return func(oldObj, obj interface{}) {
		log.Logger.Debug("Update:", oldObj, obj)

		uOldObj := oldObj.(*unstructured.Unstructured)
		oldLabels := uOldObj.GetLabels()

		uObj := obj.(*unstructured.Unstructured)
		appName := uObj.GetName()
		// example label: iter8.tools/autox-group=hello
		labels := uObj.GetLabels()

		if autoxLabelName, ok := labels[autoxLabel]; ok && appName == autoxLabelName {
			hasOldLabel := oldLabels[autoxLabel] == appName
			hasLabel := labels[autoxLabel] == appName

			// Release Helm charts
			if !hasOldLabel && hasLabel {
				iterateCharts(appName, cgc, releaseHelmChart)

				// Delete Helm charts
			} else if hasOldLabel && !hasLabel {
				iterateCharts(appName, cgc, deleteHelmChart)
			}
		}
	}
}

var deleteObject = func(cgc chartGroupConfig) func(obj interface{}) {
	return func(obj interface{}) {
		log.Logger.Debug("Delete:", obj)

		uObj := obj.(*unstructured.Unstructured)
		appName := uObj.GetName()
		// example label: iter8.tools/autox-group=hello
		labels := uObj.GetLabels()

		// check if the app name matches the name in the autox label
		if autoxLabelName, ok := labels[autoxLabel]; ok && appName == autoxLabelName {
			// Delete Helm charts
			iterateCharts(appName, cgc, deleteHelmChart)
		}
	}
}

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher(resourceTypes []schema.GroupVersionResource, namespaces []string, cgc chartGroupConfig) *iter8Watcher {
	w := &iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}
	// for each namespace, resource type configure Informer
	for _, ns := range namespaces {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sClient.dynamicClient, 0, ns, nil)
		for _, gvr := range resourceTypes {
			informer := w.factories[ns].ForResource(gvr)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    addObject(cgc),
				UpdateFunc: updateObject(cgc),
				DeleteFunc: deleteObject(cgc),
			})
		}
	}
	return w
}

func (watcher *iter8Watcher) start(stopChannel chan struct{}) {
	for _, f := range watcher.factories {
		f.Start(stopChannel)
	}
}
