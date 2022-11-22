package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"hash/maphash"
	"sync"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	autoXLabel   = "iter8.tools/autox-group"
	appLabel     = "app.kubernetes.io/name"
	versionLabel = "app.kubernetes.io/version"
	trackLabel   = "iter8.tools/track"
)

var hasher maphash.Hash
var m sync.Mutex

//go:embed application.tpl
var tplStr string

type chartAction int64

const (
	releaseAction chartAction = 0
	deleteAction  chartAction = 1
)

// applicationValues is the values for the application template
type applicationValues struct {
	// name is the name of the application
	name string

	// name is the namespace of the application
	namespace string

	// owner is the spec group secret for this application
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
//	the name of the releaseSpec,
//	the ID of the spec within the releaseSpec, and
//	the set of (pruned) labels that triggers this release
func getReleaseName(group string, releaseSpecID string, prunedLabels map[string]string) string {
	hasher.Reset()

	// use labels relevant to autoX to create a random hash value
	// this value will be appended as a suffix in the release name
	// specGroupName and specID are always hashed
	_, _ = hasher.WriteString(group)
	_, _ = hasher.WriteString(releaseSpecID)

	// hash app label
	app := prunedLabels[appLabel]
	_, _ = hasher.WriteString(app)

	// hash version label
	version := prunedLabels[versionLabel]
	_, _ = hasher.WriteString(version)

	// hash track label
	track := prunedLabels[trackLabel]
	_, _ = hasher.WriteString(track)

	nonce := fmt.Sprintf("%05x", hasher.Sum64())
	nonce = nonce[:5]
	return fmt.Sprintf("autox-%s-%s-%s", group, releaseSpecID, nonce)
}

// installHelmRelease for a given spec within a spec group
var installHelmRelease = func(releaseName string, group string, releaseSpec releaseSpec, namespace string) error {
	secretsClient := k8sClient.clientset.CoreV1().Secrets(namespace)

	// TODO: what to put for ctx?
	// get secret, based on autoX label
	labelSelector := fmt.Sprintf("%s=%s", autoXLabel, group)
	secretList, err := secretsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Logger.Error("could not list secrets")
		return err
	}

	// ensure that only one secret was found
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
			uid:  string(secret.GetUID()),
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

	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}

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

	log.Logger.Debug("application manifest:", buf.String())

	// serialize application manifest into unstructured object
	obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(buf.Bytes(), nil, nil)
	if err != nil {
		log.Logger.Error("could not serialize application manifest")
		return err
	}
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		log.Logger.Error("could not convert application manifest into unstructured object")
		return err
	}
	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

	// TODO: what to put for ctx?
	// create application object
	_, err = k8sClient.dynamic().Resource(gvr).Namespace(namespace).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		log.Logger.Error("could not create application:", releaseName)
		return err
	}

	log.Logger.Debug("Release chart:", releaseName)
	return nil
}

// deleteHelmRelease with a given release name
var deleteHelmRelease = func(releaseName string, group string, namespace string) error {
	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}

	err := k8sClient.dynamic().Resource(gvr).Namespace(namespace).Delete(context.TODO(), releaseName, metav1.DeleteOptions{})
	if err != nil {
		log.Logger.Error("could not delete application:", releaseName)
		return err
	}

	log.Logger.Debug("Delete chart:", releaseName)
	return nil
}

// doChartAction iterates through a given spec group, and performs action for each spec
// action can be install or delete
func doChartAction(prunedLabels map[string]string, chartAction chartAction, namespace string, releaseGroupSpec releaseGroupSpec) error {
	// get group
	group := prunedLabels[autoXLabel]

	var err error
	for releaseSpecID, releaseSpec := range releaseGroupSpec.ReleaseSpecs {
		// get release name
		releaseName := getReleaseName(group, releaseSpecID, prunedLabels)
		// perform action for this release
		switch chartAction {
		case releaseAction:
			// if there is an error, keep going forward in the for loop
			if err1 := installHelmRelease(releaseName, group, releaseSpec, namespace); err1 != nil {
				err = errors.New("one or more Helm release installs failed")
			}
		case deleteAction:
			// if there is an error, keep going forward in the for loop
			if err1 := deleteHelmRelease(releaseName, group, namespace); err1 != nil {
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
func pruneLabels(labels map[string]string) map[string]string {
	prunedLabels := map[string]string{}
	for _, l := range []string{autoXLabel, appLabel, versionLabel, trackLabel} {
		prunedLabels[l] = labels[l]
	}
	return prunedLabels
}

// hasAutoXLabel checks if autoX label is present
func hasAutoXLabel(labels map[string]string) bool {
	_, ok := labels[autoXLabel]
	return ok
}

// handle is the entry point to all (add, update, delete) event handlers
func handle(obj interface{}, releaseGroupSpec releaseGroupSpec) {
	m.Lock()
	defer m.Unlock()

	// parse object
	u := obj.(*unstructured.Unstructured)
	name := u.GetName()
	ns := u.GetNamespace()
	// Note: GVR is from the release group spec, not available through the obj
	gvr := getGVR(releaseGroupSpec)
	labels := u.GetLabels()
	prunedLabels := pruneLabels(labels)

	// always delete Helm releases
	_ = doChartAction(prunedLabels, deleteAction, ns, releaseGroupSpec)

	// install Helm releases if (client) object exists and has autoX label
	// fetch (client) object from cluster
	clientU, _ := k8sClient.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if clientU != nil {
		clientName := clientU.GetName()
		clientNs := clientU.GetNamespace()

		// sanity check
		if clientName != name {
			log.Logger.Errorf("autoX expected Kubernetes object to have name \"%s\" but had name \"%s\" instead", name, clientName)
			return
		}
		if clientNs != ns {
			log.Logger.Errorf("autoX expected Kubernetes object to have name \"%s\" but had name \"%s\" instead", ns, clientNs)
			return
		}

		// check if autoX label exists
		clientLabels := clientU.GetLabels()
		if !hasAutoXLabel(clientLabels) {
			return
		}

		// only install Helm releases if autoX label exists
		clientPrunedLabels := pruneLabels(clientLabels)
		_ = doChartAction(clientPrunedLabels, releaseAction, clientNs, releaseGroupSpec)
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
		handle(obj, releaseGroupSpec)
	}
}

func updateObject(releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) func(oldObj, obj interface{}) {
	return func(oldObj, obj interface{}) {
		log.Logger.Debug("Update:", obj)
		handle(obj, releaseGroupSpec)
	}
}

func deleteObject(releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) func(obj interface{}) {
	return func(obj interface{}) {
		log.Logger.Debug("Delete:", obj)
		handle(obj, releaseGroupSpec)
	}
}

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher(autoXConfig config) *iter8Watcher {
	w := &iter8Watcher{
		// the key is releaseGroupSpecName
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}

	// aggregate all triggers (namespaces and GVR) from the releaseGroupConfig
	// triggers map has namespace as its key and the object GVRs within the namespace that it is watching as its value
	// triggers := map[string][]schema.GroupVersionResource{}
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
