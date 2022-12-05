package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// this label is used in trigger objects
	// label always set to true when present
	// AutoX controller will only check for the existence of this label, not its value
	autoXLabel = "iter8.tools/autox"

	// this label is used in secrets (to allow for ownership over the application objects)
	// label is set to the name of a release group spec (releaseGroupSpecName)
	// there is a 1:1 mapping of secrets to release group specs
	autoXGroupLabel = "iter8.tools/autox-group"

	autoXAdditionValues = "autoXAdditionalValues"
	appLabel            = "app.kubernetes.io/name"
	versionLabel        = "app.kubernetes.io/version"
	trackLabel          = "iter8.tools/track"
)

var m sync.Mutex

//go:embed application.tpl
var tplStr string

type chartAction int64

const (
	applyAction  chartAction = 0
	deleteAction chartAction = 1
)

// applicationValues is the values for the (Argo CD) application template
type applicationValues struct {
	// name is the name of the application
	name string

	// name is the namespace of the application
	namespace string

	// owner is the release group spec secret for this application
	// we create an secret for each release group spec
	// this secret is assigned as the owner of this spec
	// when we delete the secret, the application is also deleted
	owner struct {
		name string
		uid  string
	}

	// chart is the Helm chart for this application
	chart struct {
		name    string
		values  map[string]interface{}
		version string
	}
}

// the name of a release will depend on:
//
//	the name of the release group spec (releaseGroupSpecName)
//	the ID of the release spec (releaseSpecID)
func getReleaseName(releaseGroupSpecName string, releaseSpecID string) string {
	return fmt.Sprintf("autox-%s-%s", releaseGroupSpecName, releaseSpecID)
}

// applyApplicationObject will apply an application based on a release spec
var applyApplicationObject = func(releaseName string, releaseGroupSpecName string, releaseSpec releaseSpec, namespace string, additionalValues map[string]string) error {
	secretsClient := k8sClient.clientset.CoreV1().Secrets(namespace)

	// get secret, based on autoX group label
	labelSelector := fmt.Sprintf("%s=%s", autoXGroupLabel, releaseGroupSpecName)
	secretList, err := secretsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Logger.Error("could not list secrets")
		return err
	}

	// ensure that only one secret is found
	if secretsLen := len(secretList.Items); secretsLen == 0 {
		log.Logger.Error("expected secret with label selector:", labelSelector, "but none were found")
		return err
	} else if secretsLen > 1 {
		log.Logger.Error("expected secret with label selector:", labelSelector, "but more than one were found")
		return err
	}
	secret := secretList.Items[0]

	values := applicationValues{
		name:      releaseName,
		namespace: namespace,

		owner: struct {
			name string
			uid  string
		}{
			name: secret.Name,
			uid:  string(secret.GetUID()), // assign the release group spec secret as the owner of the application
		},

		chart: struct {
			name    string
			values  map[string]interface{}
			version string
		}{
			name:    releaseSpec.Name,
			values:  releaseSpec.Values,
			version: releaseSpec.Version,
		},
	}

	// add additionalValues to the values
	// Argo CD will create a new experiment if it sees that the additionalValues are different from the previous experiment
	// additionalValues will contain the pruned labels from the Kubernetes object
	values.chart.values[autoXAdditionValues] = additionalValues

	tpl, err := base.CreateTemplate(tplStr)
	if err != nil {
		log.Logger.Error("could not create application template")
		return err
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, values)
	if err != nil {
		log.Logger.Error("could not execute application template")
		return err
	}

	// decode object into unstructured.UnstructuredJSONScheme
	// source: https://github.com/kubernetes/client-go/blob/1ac8d459351e21458fd1041f41e43403eadcbdba/dynamic/simple.go#L186
	uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, buf.Bytes())
	if err != nil {
		log.Logger.Error("could not decode object into unstructured.UnstructuredJSONScheme:", buf.String())
		return err
	}

	// apply application object to the K8s cluster
	log.Logger.Debug("apply application object:", releaseName)
	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	_, err = k8sClient.dynamic().Resource(gvr).Namespace(namespace).Apply(context.TODO(), releaseName, uncastObj.(*unstructured.Unstructured), metav1.ApplyOptions{})
	if err != nil {
		log.Logger.Error("could not create application:", releaseName)
		return err
	}

	return nil
}

// deleteApplicationObject deletes an application object based on a given release name
var deleteApplicationObject = func(releaseName string, releaseGroupSpecName string, namespace string) error {
	log.Logger.Debug("delete application object:", releaseName)

	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	err := k8sClient.dynamic().Resource(gvr).Namespace(namespace).Delete(context.TODO(), releaseName, metav1.DeleteOptions{})
	if err != nil {
		log.Logger.Error("could not delete application:", releaseName)
		return err
	}

	return nil
}

// doChartAction iterates through a release group spec and performs apply/delete action for each release spec
// action can be apply or delete
func doChartAction(chartAction chartAction, releaseGroupSpecName string, namespace string, releaseGroupSpec releaseGroupSpec, additionalValues map[string]string) error {
	// get group
	var err error
	for releaseSpecID, releaseSpec := range releaseGroupSpec.ReleaseSpecs {
		// get release name
		releaseName := getReleaseName(releaseGroupSpecName, releaseSpecID)

		// perform action for this release
		switch chartAction {
		case applyAction:
			// if there is an error, keep going forward in the for loop
			if err1 := applyApplicationObject(releaseName, releaseGroupSpecName, releaseSpec, namespace, additionalValues); err1 != nil {
				err = errors.New("one or more Helm release applys failed")
			}

		case deleteAction:
			// if there is an error, keep going forward in the for loop
			if err1 := deleteApplicationObject(releaseName, releaseGroupSpecName, namespace); err1 != nil {
				err = errors.New("one or more Helm release deletions failed")
			}
		}
	}

	if err != nil {
		log.Logger.Error(err)
	}

	return err
}

// pruneLabels will extract the labels that are relevant for autoX
// currently, the important labels are:
//
//	autoXLabel   = "iter8.tools/autox"
//	appLabel     = "app.kubernetes.io/name"
//	versionLabel = "app.kubernetes.io/version"
//	trackLabel   = "iter8.tools/track"
func pruneLabels(labels map[string]string) map[string]string {
	prunedLabels := map[string]string{}
	for _, l := range []string{autoXLabel, appLabel, versionLabel, trackLabel} {
		prunedLabels[l] = labels[l]
	}
	return prunedLabels
}

// hasAutoXLabel checks if autoX label is present
// autoX label is used to determine if any autoX functionality should be performed
func hasAutoXLabel(labels map[string]string) bool {
	_, ok := labels[autoXLabel]
	return ok
}

// handle is the entry point to all (add, update, delete) event handlers
func handle(obj interface{}, releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) {
	m.Lock()
	defer m.Unlock()

	// parse object
	u := obj.(*unstructured.Unstructured)

	// check if name matches trigger
	name := u.GetName()
	if name != releaseGroupSpec.Trigger.Name {
		return
	}

	// at this point, we know that we are really handling an even for the trigger object
	// name, namespace, and GVR should all match

	// namespace and GVR should already match trigger
	ns := u.GetNamespace()
	// Note: GVR is from the release group spec, not available through the obj
	gvr := getGVR(releaseGroupSpec)

	// get (client) object from cluster
	clientU, _ := k8sClient.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})

	// if (client) object exists
	// delete application objects if (client) object does not have autoX label
	// then apply application objects if (client) object has autoX label
	if clientU != nil {
		// check if autoX label does not exist
		clientLabels := clientU.GetLabels()
		if !hasAutoXLabel(clientLabels) {
			log.Logger.Debugf("delete application objects for release group \"%s\"", releaseGroupSpecName)

			_ = doChartAction(deleteAction, releaseGroupSpecName, ns, releaseGroupSpec, nil)

			// if autoX label does not exist, there is no need to apply application objects, so return
			return
		}

		// apply application objects for the release group
		log.Logger.Debugf("apply application objects for release group \"%s\"", releaseGroupSpecName)
		clientPrunedLabels := pruneLabels(clientLabels)
		_ = doChartAction(applyAction, releaseGroupSpecName, ns, releaseGroupSpec, clientPrunedLabels)

		// delete application objects if (client) object does not exist
	} else {
		log.Logger.Debugf("delete application objects for release group \"%s\"", releaseGroupSpecName)

		_ = doChartAction(deleteAction, releaseGroupSpecName, ns, releaseGroupSpec, nil)
	}
}

// getGVR gets the namespace and GVR from a release group spec trigger
func getGVR(releaseGroupSpec releaseGroupSpec) schema.GroupVersionResource {
	gvr := schema.GroupVersionResource{
		Group:    releaseGroupSpec.Trigger.Group,
		Version:  releaseGroupSpec.Trigger.Version,
		Resource: releaseGroupSpec.Trigger.Resource,
	}

	return gvr
}

func addObject(releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) func(obj interface{}) {
	return func(obj interface{}) {
		log.Logger.Debug("Add:", obj)
		handle(obj, releaseGroupSpecName, releaseGroupSpec)
	}
}

func updateObject(releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) func(oldObj, obj interface{}) {
	return func(oldObj, obj interface{}) {
		log.Logger.Debug("Update:", obj)
		handle(obj, releaseGroupSpecName, releaseGroupSpec)
	}
}

func deleteObject(releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) func(obj interface{}) {
	return func(obj interface{}) {
		log.Logger.Debug("Delete:", obj)
		handle(obj, releaseGroupSpecName, releaseGroupSpec)
	}
}

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher(autoXConfig config) *iter8Watcher {
	w := &iter8Watcher{
		// the key is the name of the release group spec (releaseGroupSpecName)
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}

	// create a factory for each trigger
	// there is a 1:1 correspondence between each trigger and release group spec
	// effectively, we are creating one factory per trigger
	// the key to the factories map is the name of the release group spec (releaseGroupSpecName)
	for releaseGroupSpecName, releaseGroupSpec := range autoXConfig.Specs {
		ns := releaseGroupSpec.Trigger.Namespace
		gvr := getGVR(releaseGroupSpec)

		w.factories[releaseGroupSpecName] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sClient.dynamicClient, 0, ns, nil)

		informer := w.factories[releaseGroupSpecName].ForResource(gvr)
		informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    addObject(releaseGroupSpecName, releaseGroupSpec),
			UpdateFunc: updateObject(releaseGroupSpecName, releaseGroupSpec),
			DeleteFunc: deleteObject(releaseGroupSpecName, releaseGroupSpec),
		})
	}

	return w
}

func (watcher *iter8Watcher) start(stopChannel chan struct{}) {
	for _, f := range watcher.factories {
		f.Start(stopChannel)
	}
}
