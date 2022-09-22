package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"errors"
	"fmt"
	"hash/maphash"
	"reflect"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

const (
	autoXLabel = "iter8.tools/autox-group"
)

type chartAction int64

const (
	releaseAction chartAction = 0
	deleteAction  chartAction = 1
)

// the name of a release will depend on:
//
//	the name of the releaseSpec,
//	the id of the chart within the releaseSpec, and
//	the set of (pruned) labels that triggers this release
func getReleaseName(chartGroupName string, chartID string, prunedLabels map[string]string) string {

	// use labels relevant to autoX to create a random hash value
	// this value will be appended as a suffix in the release name
	var hasher maphash.Hash
	// chartGroupName and chartID are always hashed
	_, _ = hasher.WriteString(chartGroupName)
	_, _ = hasher.WriteString(chartID)

	// adding version label
	version := prunedLabels["app.kubernetes.io/version"]
	_, _ = hasher.WriteString(version)

	// adding track label
	track := prunedLabels["iter8.tools/track"]
	_, _ = hasher.WriteString(track)

	nonce := fmt.Sprintf("%05x", hasher.Sum64())
	nonce = nonce[:5]
	return fmt.Sprintf("autox-%s-%s-%s", chartGroupName, chartID, nonce)
}

// installHelmReleases for a given chart group
func installHelmReleases(prunedLabels map[string]string, namespace string) error {
	return doChartAction(prunedLabels, releaseAction, namespace)
}

// installHelmRelease for a given chart within a chart group
func installHelmRelease(releaseName string, chart chart, namespace string) error {
	// TODO: check if there is a preexisting Helm release

	// TODO: mutex

	// TODO: release Helm chart

	log.Logger.Debug("Release chart:", releaseName)
	return nil
}

// deleteHelmReleases for a given chart group
func deleteHelmReleases(prunedLabels map[string]string, namespace string) error {
	return doChartAction(prunedLabels, deleteAction, namespace)
}

// deleteHelmRelease with a given release name
func deleteHelmRelease(releaseName string, namespace string) error {
	// TODO: check if there is a preexisting Helm release

	// TODO: mutex

	// TODO: delete Helm chart

	log.Logger.Debug("Delete chart:", releaseName)
	return nil
}

// doChartAction iterates through a given chart group, and performs action for each chart
// action can be install or delete
func doChartAction(prunedLabels map[string]string, chartAction chartAction, namespace string) error {
	// get chart group name
	chartGroupName := prunedLabels[autoXLabel]
	// iterate through the charts in this chart group
	var err error
	if cg, ok := iter8ChartGroupConfig[chartGroupName]; ok {
		for chartID, chart := range cg.Charts {
			// get release name
			releaseName := getReleaseName(chartGroupName, chartID, prunedLabels)
			// perform action for this release
			switch chartAction {
			case releaseAction:
				// if there is an error, keep going forward in the for loop
				if err1 := installHelmRelease(releaseName, chart, namespace); err1 != nil {
					err = errors.New("one or more Helm release installs failed")
				}
			case deleteAction:
				// if there is an error, keep going forward in the for loop
				if err1 := deleteHelmRelease(releaseName, namespace); err1 != nil {
					err = errors.New("one or more Helm release deletions failed")
				}
			}
		}
	} else {
		log.Logger.Warnf("no matching chart group name in autoX group configuration: %s", chartGroupName)
	}

	if err != nil {
		log.Logger.Error(err)
	}

	return err
}

// addObject is the function object that will be used as the add handler in the informer
var addObject = func(obj interface{}) {
	log.Logger.Debug("Add:", obj)

	uObj := obj.(*unstructured.Unstructured)

	// if there is no autoX label, there is nothing to do
	labels := uObj.GetLabels()
	if !gotAutoXLabel(labels) {
		return
	}
	// there is an autoX group name

	// we will install Helm releases
	prunedLabels := pruneLabels(labels)
	installHelmReleases(prunedLabels, uObj.GetNamespace())
}

// pruneLabels will extract the labels that are relevant for autoX
func pruneLabels(labels map[string]string) map[string]string {
	prunedLabels := map[string]string{}
	// if the autoX label exists, get it and stuff it into pruneLabels
	if autoXLabelValue, ok := labels[autoXLabel]; ok {
		prunedLabels[autoXLabel] = autoXLabelValue
	}
	// other labels like track and version can be potentially extracted in the future
	return prunedLabels
}

// gotAutoXLabel checks if autoX label is present
func gotAutoXLabel(labels map[string]string) bool {
	_, ok := labels[autoXLabel]
	return ok
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
	labels := pruneLabels(uObj.GetLabels())

	// check if the app name matches the name in the autox label
	if autoxLabelName, ok := labels[autoxLabel]; ok && resourceName == autoxLabelName {
		// Delete Helm charts
		deleteHelmReleases(resourceName)
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
