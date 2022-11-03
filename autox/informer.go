package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"hash/maphash"
	"reflect"
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
		url     string
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

	// use labels relevant to autoX to create a random hash value
	// this value will be appended as a suffix in the release name
	var hasher maphash.Hash
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

// installHelmReleases for a given spec group
func installHelmReleases(prunedLabels map[string]string, namespace string) error {
	return doChartAction(prunedLabels, releaseAction, namespace)
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
			url     string
			name    string
			values  map[string]interface{}
			version string
		}{
			url:     releaseSpec.RepoURL,
			name:    releaseSpec.Name,
			values:  releaseSpec.Values,
			version: releaseSpec.Version,
		},
	}

	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

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

// deleteHelmReleases for a given spec group
func deleteHelmReleases(prunedLabels map[string]string, namespace string) error {
	return doChartAction(prunedLabels, deleteAction, namespace)
}

// deleteHelmRelease with a given release name
var deleteHelmRelease = func(releaseName string, group string, namespace string) error {
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

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
func doChartAction(prunedLabels map[string]string, chartAction chartAction, namespace string) error {
	// get group
	group := prunedLabels[autoXLabel]

	// iterate through the specs in this spec group
	var err error
	if releaseGroupSpec, ok := autoXConfig.Specs[group]; ok {
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
	} else {
		log.Logger.Warnf("no matching group name in autoX group configuration: %s", group)
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

// addObject is the function object that will be used as the add handler in the informer
func addObject(ns string, gvr schema.GroupVersionResource, m *sync.Mutex) func(obj interface{}) {
	return func(obj interface{}) {
		m.Lock()
		defer m.Unlock()

		log.Logger.Debug("Add:", obj)

		u := obj.(*unstructured.Unstructured)
		clientU, err := k8sClient.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), u.GetName(), metav1.GetOptions{})
		if err != nil {
			log.Logger.Warn(err)
			log.Logger.Warnf("could not get object \"%v\" from client", u)
			return
		}

		if clientU == nil {
			log.Logger.Warnf("expected object \"%v\" to exist but none were found", u)
			return
		}

		// check if autoX label exists
		labels := clientU.GetLabels()
		if !hasAutoXLabel(labels) {
			log.Logger.Warnf("expected object \"%v\" to have label \"%s\" but none were found", clientU, autoXLabel)
			return
		}

		// only install Helm releases if auto X label exists
		prunedLabels := pruneLabels(labels)
		_ = addObjectHelper(clientU, prunedLabels)
	}
}

func addObjectHelper(uObj *unstructured.Unstructured, prunedLabels map[string]string) error {
	return installHelmReleases(prunedLabels, uObj.GetNamespace())
}

func updateObject(ns string, gvr schema.GroupVersionResource, m *sync.Mutex) func(oldObj, obj interface{}) {
	return func(oldObj, obj interface{}) {
		m.Lock()
		defer m.Unlock()

		log.Logger.Debug("Update:", oldObj, obj)

		oldU := oldObj.(*unstructured.Unstructured)
		oldPrunedLabels := pruneLabels(oldU.GetLabels())

		u := obj.(*unstructured.Unstructured)
		clientU, err := k8sClient.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), u.GetName(), metav1.GetOptions{})
		if err != nil {
			log.Logger.Warn(err)
			log.Logger.Warnf("could not get object \"%v\" from client", u)
			return
		}

		// TODO: if we cannot get the object from the dynamicClient, should we send a warning and stop or should we
		// skip the Helm deletion and go directly to Helm install?
		if clientU == nil {
			log.Logger.Warnf("expected object \"%v\" to exist but none were found", u)
			return
		}

		prunedLabels := pruneLabels(clientU.GetLabels())

		// check if labels have changed
		if reflect.DeepEqual(oldPrunedLabels, prunedLabels) {
			log.Logger.Debug("updated object has no labels changes")
			return
		}

		// if labels have changed, then delete the Helm releases
		_ = deleteObjectHelper(clientU, prunedLabels)

		// check if autoX label exists
		labels := clientU.GetLabels()
		if !hasAutoXLabel(labels) {
			log.Logger.Warnf("expected object \"%v\" to have label \"%s\" but none were found", clientU, autoXLabel)
			return
		}

		// only install Helm releases if auto X label exists
		_ = addObjectHelper(clientU, prunedLabels)
	}
}

func deleteObject(ns string, gvr schema.GroupVersionResource, m *sync.Mutex) func(obj interface{}) {
	return func(obj interface{}) {
		m.Lock()
		defer m.Unlock()

		log.Logger.Debug("Delete:", obj)

		u := obj.(*unstructured.Unstructured)
		clientU, _ := k8sClient.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), u.GetName(), metav1.GetOptions{})
		if clientU != nil {
			log.Logger.Warnf("object \"%v\" should have been deleted but was not", clientU)
			return
		}

		// delete Helm releases if auto X label exists
		labels := u.GetLabels()
		prunedLabels := pruneLabels(labels)
		_ = deleteObjectHelper(u, prunedLabels)
	}
}

func deleteObjectHelper(uObj *unstructured.Unstructured, prunedLabels map[string]string) error {
	return deleteHelmReleases(prunedLabels, uObj.GetNamespace())
}

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher() *iter8Watcher {
	w := &iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}

	// aggregate all triggers (namespaces and GVR) from the releaseGroupConfig
	// triggers map has namespace as its key and the object GVRs within the namespace that it is watching as its value
	triggers := map[string][]schema.GroupVersionResource{}
	for _, releaseGroupSpec := range autoXConfig.Specs {

		namespace := releaseGroupSpec.Trigger.Namespace
		gvr := schema.GroupVersionResource{
			Group:    releaseGroupSpec.Trigger.Group,
			Version:  releaseGroupSpec.Trigger.Version,
			Resource: releaseGroupSpec.Trigger.Resource,
		}

		// add namespace and GVR to triggers
		triggers[namespace] = append(triggers[namespace], gvr)
	}

	// mutex is passed to event handler so changes are not concurrent
	var m sync.Mutex

	// for each namespace and resource type, configure an informer
	for ns, gvrs := range triggers {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sClient.dynamicClient, 0, ns, nil)

		// TBD: optimize the informers by supplying the trigger object name, perhaps through list options
		for _, gvr := range gvrs {
			informer := w.factories[ns].ForResource(gvr)
			informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc:    addObject(ns, gvr, &m),
				UpdateFunc: updateObject(ns, gvr, &m),
				DeleteFunc: deleteObject(ns, gvr, &m),
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
