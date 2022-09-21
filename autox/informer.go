package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"fmt"
	"reflect"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

const (
	autoxLabel = "iter8.tools/autox-group"
)

type chartAction int64

const (
	releaseAction chartAction = 0
	deleteAction  chartAction = 1
)

func getReleaseName(chartGroupName string, chartName string) string {
	return fmt.Sprintf("autox-%s-%s", chartGroupName, chartName)
}

func installHelmReleases(resourceName string) {
	doChartAction(resourceName, releaseAction)
}

func installHelmRelease(releaseName string, chart chart) {
	// TODO: check if there is a preexisting Helm release

	// TODO: mutex

	// TODO: release Helm chart

	log.Logger.Debug("Release chart:", releaseName)
}

func deleteHelmReleases(resourceName string) {
	doChartAction(resourceName, deleteAction)
}

func deleteHelmRelease(releaseName string, chart chart) {
	// TODO: check if there is a preexisting Helm release

	// TODO: mutex

	// TODO: delete Helm chart

	log.Logger.Debug("Delete chart:", releaseName)
}

func doChartAction(resourceName string, chartAction chartAction) {
	if cg, ok := iter8ChartGroupConfig[resourceName]; ok {
		for chartName, chart := range cg.Charts {
			releaseName := getReleaseName(resourceName, chartName)

			switch chartAction {
			case releaseAction:
				installHelmRelease(releaseName, chart)

			case deleteAction:
				deleteHelmRelease(releaseName, chart)
			}
		}
	} else {
		// TODO: what log level should this be?
		log.Logger.Debug("AutoX should make a Helm release for app \"", resourceName, "\" but no Helm charts were provided in the chartGroupConfig")
	}
}

var addObject = func(obj interface{}) {
	log.Logger.Debug("Add:", obj)

	uObj := obj.(*unstructured.Unstructured)
	resourceName := uObj.GetName()
	// example label: iter8.tools/autox-group=hello
	labels := uObj.GetLabels()

	// check if the app name matches the name in the autox label
	if autoxLabelName, ok := labels[autoxLabel]; ok && resourceName == autoxLabelName {
		// Release Helm charts
		doChartAction(resourceName, releaseAction)
	}
}

func pruneLabels(labels map[string]string) map[string]string {
	// TODO: select labels important for autoX

	return labels
}

var updateObject = func(oldObj, obj interface{}) {
	log.Logger.Debug("Update:", oldObj, obj)

	uOldObj := oldObj.(*unstructured.Unstructured)
	oldLabels := pruneLabels(uOldObj.GetLabels())

	uObj := obj.(*unstructured.Unstructured)
	resourceName := uObj.GetName()
	// example label: iter8.tools/autox-group=hello
	labels := pruneLabels(uObj.GetLabels())

	// if the pruned labels are the same, do nothing
	if reflect.DeepEqual(oldLabels, labels) {
		return
	}

	// if the pruned labels are different, then update by deleting and reinstalling
	deleteHelmReleases(resourceName)
	installHelmReleases(resourceName)
}

var deleteObject = func(obj interface{}) {
	log.Logger.Debug("Delete:", obj)

	uObj := obj.(*unstructured.Unstructured)
	resourceName := uObj.GetName()
	// example label: iter8.tools/autox-group=hello
	labels := uObj.GetLabels()

	// check if the app name matches the name in the autox label
	if autoxLabelName, ok := labels[autoxLabel]; ok && resourceName == autoxLabelName {
		// Delete Helm charts
		doChartAction(resourceName, deleteAction)
	}
}

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher() *iter8Watcher {
	w := &iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}
	// for each namespace, resource type configure Informer
	for _, ns := range iter8ResourceConfig.Namespaces {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sClient.dynamicClient, 0, ns, nil)
		for _, gvr := range iter8ResourceConfig.Resources {
			informer := w.factories[ns].ForResource(gvr)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    addObject,
				UpdateFunc: updateObject,
				DeleteFunc: deleteObject,
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
