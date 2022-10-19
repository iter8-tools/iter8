package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/maphash"
	"os"
	"reflect"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	autoXLabel   = "iter8.tools/autox-group"
	appLabel     = "app.kubernetes.io/name"
	versionLabel = "app.kubernetes.io/version"
	trackLabel   = "iter8.tools/track"
)

type chartAction int64

const (
	releaseAction chartAction = 0
	deleteAction  chartAction = 1

	applicationTemplateFilePath string = "application.tpl"
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
	secretList, err := secretsClient.List(context.TODO(), metaV1.ListOptions{
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

	// get application template
	dat, err := os.ReadFile(applicationTemplateFilePath)
	if err != nil {
		log.Logger.Error("could not read application template")
		return err
	}

	tpl, err := base.CreateTemplate(string(dat))
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
	_, err = k8sClient.dynamic().Resource(gvr).Namespace(namespace).Create(context.TODO(), unstructuredObj, metaV1.CreateOptions{})
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

	err := k8sClient.dynamic().Resource(gvr).Namespace(namespace).Delete(context.TODO(), releaseName, metaV1.DeleteOptions{})
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
func addObject(obj interface{}) {
	log.Logger.Debug("Add:", obj)

	uObj := obj.(*unstructured.Unstructured)

	// if there is no autoX label, there is nothing to do
	labels := uObj.GetLabels()
	if !hasAutoXLabel(labels) {
		return
	}
	// there is an autoX group name

	// we will install Helm releases
	prunedLabels := pruneLabels(labels)

	_ = installHelmReleases(prunedLabels, uObj.GetNamespace())
}

func updateObject(oldObj, obj interface{}) {
	log.Logger.Debug("Update:", oldObj, obj)

	uOldObj := oldObj.(*unstructured.Unstructured)
	prunedLabelsOld := pruneLabels(uOldObj.GetLabels())

	uObj := obj.(*unstructured.Unstructured)
	prunedLabels := pruneLabels(uObj.GetLabels())

	// if the pruned label sets are equal, do nothing
	if reflect.DeepEqual(prunedLabelsOld, prunedLabels) {
		return
	}

	// if the pruned label sets are different, then
	// first attempt delete, and then attempt install
	if hasAutoXLabel(prunedLabelsOld) {
		_ = deleteHelmReleases(prunedLabelsOld, uOldObj.GetNamespace())
	}

	if hasAutoXLabel(prunedLabels) {
		_ = installHelmReleases(prunedLabels, uOldObj.GetNamespace())
	}
}

func deleteObject(obj interface{}) {
	log.Logger.Debug("Delete:", obj)

	uObj := obj.(*unstructured.Unstructured)
	prunedLabels := pruneLabels(uObj.GetLabels())

	if !hasAutoXLabel(prunedLabels) {
		return
	}

	// delete Helm charts
	_ = deleteHelmReleases(prunedLabels, uObj.GetNamespace())
}

type iter8Watcher struct {
	factories map[string]dynamicinformer.DynamicSharedInformerFactory
}

func newIter8Watcher(k8sClient *kubeClient) *iter8Watcher {
	w := &iter8Watcher{
		factories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
	}

	// aggregate all triggers (namespaces and GVR) from the releaseGroupConfig
	triggers := map[string][]schema.GroupVersionResource{}
	for _, releaseGroupSpec := range autoXConfig.Specs {

		namespace := releaseGroupSpec.Trigger.Namespace
		gvr := schema.GroupVersionResource{
			Group:    releaseGroupSpec.Trigger.Group,
			Version:  releaseGroupSpec.Trigger.Version,
			Resource: releaseGroupSpec.Trigger.Resource,
		}

		// add namespace and GVR to triggers
		if _, ok := triggers[namespace]; ok {
			triggers[namespace] = []schema.GroupVersionResource{gvr}
		} else {
			triggers[namespace] = append(triggers[namespace], gvr)
		}
	}

	// for each namespace, resource type configure Informer
	for ns, gvrs := range triggers {
		w.factories[ns] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sClient.dynamicClient, 0, ns, nil)
		for _, gvr := range gvrs {
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
