package autox

// informer.go - informer(s) to watch desired resources/namespaces

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"

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

	argocd = "argocd"

	templateRevision = "templateRevision"
	experimentYAML   = "experiment.yaml"
)

var applicationGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

var m sync.Mutex

//go:embed application.tpl
var tplStr string

type chartAction int64

const (
	applyAction  chartAction = 0
	deleteAction chartAction = 1
)

type owner struct {
	Name string `json:"name" yaml:"name"`
	Uid  string `json:"uid" yaml:"uid"`
}

// applicationValues is the values for the (Argo CD) application template
type applicationValues struct {
	// Name is the name of the application
	Name string `json:"name" yaml:"name"`

	// Namespace is the namespace of the application
	Namespace string `json:"namespace" yaml:"namespace"`

	// Owner is the release group spec secret for this application
	// we create an secret for each release group spec
	// this secret is assigned as the Owner of this spec
	// when we delete the secret, the application is also deleted
	Owner owner `json:"owner" yaml:"owner"`

	// Chart is the Helm Chart for this application
	Chart releaseSpec `json:"chart" yaml:"chart"`
}

// the name of a release will depend on:
//
//	the name of the release group spec (releaseGroupSpecName)
//	the ID of the release spec (releaseSpecID)
func getReleaseName(releaseGroupSpecName string, releaseSpecID string) string {
	return fmt.Sprintf("autox-%s-%s", releaseGroupSpecName, releaseSpecID)
}

// shouldCreateApplication will return true if an application object should be created
// an application object should be created if the values are different from those from the previous application object (if one exists)
func shouldCreateApplication(values map[string]interface{}, releaseName string) (bool, error) {
	// get application object
	uPApp, _ := k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Get(context.TODO(), releaseName, metav1.GetOptions{}) // *unstructured.Unstructured previous application
	if uPApp != nil {
		log.Logger.Debug(fmt.Sprintf("found previous application \"%s\"", releaseName))

		pValuesString, ok, err := unstructured.NestedString(uPApp.UnstructuredContent(), "spec", "source", "helm", "values") // pValuesString previous values
		if err != nil {
			log.Logger.Error(fmt.Sprintf("cannot extract values of previous application \"%s\": %s", releaseName, pValuesString), err)
			return true, err // TODO: still return true?
		}

		// if there is values in previous application
		if ok {
			var pValues map[string]interface{}

			err = yaml.Unmarshal([]byte(pValuesString), &pValues)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("cannot parse values of previous application \"%s\": %s", releaseName, pValuesString), err)
				return true, err // TODO: still return true?
			}

			return reflect.DeepEqual(pValues, values), nil
		}

		// // convert application unstructured.Unstructured to application object
		// // See: https://erwinvaneyk.nl/kubernetes-unstructured-to-typed/
		// var pApp argo.Application // previous application
		// err = runtime.DefaultUnstructuredConverter.FromUnstructured(uPApp.UnstructuredContent(), &pApp)
		// if err != nil {
		// 	log.Logger.Error(fmt.Sprintf("cannot parse preexisting application \"%s\"", releaseName), err)
		// 	// TODO: throw error?
		// 	return false, err
		// }

		// log.Logger.Debug("parse string values: ", pApp.Spec.Source.Helm.Values)

		// // parse string values from application
		// sPValues := pApp.Spec.Source.Helm.Values // string previous (application) values
		// pValues := map[string]interface{}{}      // previous (application) values
		// err = yaml.Unmarshal([]byte(sPValues), pValues)
		// if err != nil {
		// 	log.Logger.Error(fmt.Sprintf("cannot unmarshal values from previous application \"%s\": %s", releaseName, sPValues), err)
		// 	// TODO: throw error?
		// 	return false, err
		// }

		// log.Logger.Debug("old values: ", pValues, " new values: ", values)

		// // compare values from previous application to values for the pending one
		// return reflect.DeepEqual(pValues, values), nil
	}

	// there is no preexisting application object, so should create one
	return true, nil
}

// applyApplicationObject will apply an application based on a release spec
var applyApplicationObject = func(releaseName string, releaseGroupSpecName string, releaseSpec releaseSpec, namespace string, additionalValues map[string]interface{}) error {
	secretsClient := k8sClient.clientset.CoreV1().Secrets(argocd)

	// get secret, based on autoX group label
	labelSelector := fmt.Sprintf("%s=%s", autoXGroupLabel, releaseGroupSpecName)
	secretList, err := secretsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Logger.Error("could not list release group spec secrets: ", err)
		return err
	}

	// ensure that only one secret is found
	if secretsLen := len(secretList.Items); secretsLen == 0 {
		log.Logger.Error(fmt.Sprintf("expected release group spec secret with label selector \"%s\" but none were found", labelSelector))
		return err
	} else if secretsLen > 1 {
		log.Logger.Error(fmt.Sprintf("expected release group spec secret with label selector \"%s\" but more than one were found", labelSelector))
		return err
	}
	secret := secretList.Items[0]

	tValues := applicationValues{ // template values
		Name:      releaseName,
		Namespace: namespace,

		Owner: owner{
			Name: secret.Name,
			Uid:  string(secret.GetUID()), // assign the release group spec secret as the owner of the application
		},

		Chart: releaseSpec,
	}

	// add additionalValues to the values
	// Argo CD will create a new experiment if it sees that the additionalValues are different from the previous experiment
	// additionalValues will contain the pruned labels from the Kubernetes object
	tValues.Chart.Values[autoXAdditionValues] = additionalValues

	// check if the pending application will be different from the previous application, if it exists
	// only create a new application if it will be different
	if s, _ := shouldCreateApplication(tValues.Chart.Values, releaseName); s {
		// execute application template
		tpl, err := base.CreateTemplate(tplStr)
		if err != nil {
			log.Logger.Error("could not create application template: ", err)
			return err
		}

		var buf bytes.Buffer
		err = tpl.Execute(&buf, tValues)
		if err != nil {
			log.Logger.Error("could not execute application template: ", err)
			return err
		}

		jsonBytes, err := yaml.YAMLToJSON(buf.Bytes())
		if err != nil {
			log.Logger.Error("could not convert YAML to JSON: ", buf.String(), err)
			return err
		}

		// decode pending application object into unstructured.UnstructuredJSONScheme
		// source: https://github.com/kubernetes/client-go/blob/1ac8d459351e21458fd1041f41e43403eadcbdba/dynamic/simple.go#L186
		uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsonBytes)
		if err != nil {
			log.Logger.Error("could not decode object into unstructured.UnstructuredJSONScheme: ", buf.String(), err)
			return err
		}

		// // apply application object to the K8s cluster
		// log.Logger.Debug("apply application object: ", releaseName)
		// _, err = k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Create(context.TODO(), uncastObj.(*unstructured.Unstructured), metav1.CreateOptions{})
		// if err != nil {
		// 	log.Logger.Error("could not create application: ", releaseName, err)
		// 	return err
		// }

		// apply application object to the K8s cluster
		log.Logger.Debug("apply application object: ", releaseName)
		_, err = k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Create(context.TODO(), uncastObj.(*unstructured.Unstructured), metav1.CreateOptions{})
		if err != nil {
			log.Logger.Error("could not create application: ", releaseName, err)
			return err
		}
	}

	return nil
}

// deleteApplicationObject deletes an application object based on a given release name
var deleteApplicationObject = func(releaseName string, releaseGroupSpecName string) error {
	log.Logger.Debug("delete application object: ", releaseName)

	err := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Delete(context.TODO(), releaseName, metav1.DeleteOptions{})
	if err != nil {
		log.Logger.Error("could not delete application: ", releaseName, err)
		return err
	}

	return nil
}

// doChartAction iterates through a release group spec and performs apply/delete action for each release spec
// action can be apply or delete
func doChartAction(chartAction chartAction, releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec, namespace string, additionalValues map[string]interface{}) error {
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
				err = errors.New("one or more Helm release applications failed")
			}

		case deleteAction:
			// if there is an error, keep going forward in the for loop
			if err1 := deleteApplicationObject(releaseName, releaseGroupSpecName); err1 != nil {
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
func pruneLabels(labels map[string]string) map[string]interface{} {
	prunedLabels := map[string]interface{}{}
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
	log.Logger.Debug("handle kubernetes resource object: ", obj)

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
			log.Logger.Debugf("delete application objects for release group \"%s\" (no %s label)", releaseGroupSpecName, autoXLabel)

			_ = doChartAction(deleteAction, releaseGroupSpecName, releaseGroupSpec, "", nil)

			// if autoX label does not exist, there is no need to apply application objects, so return
			return
		}

		// apply application objects for the release group
		clientPrunedLabels := pruneLabels(clientLabels)
		_ = doChartAction(applyAction, releaseGroupSpecName, releaseGroupSpec, ns, clientPrunedLabels)
	} else { // delete application objects if (client) object does not exist
		_ = doChartAction(deleteAction, releaseGroupSpecName, releaseGroupSpec, "", nil)
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
		handle(obj, releaseGroupSpecName, releaseGroupSpec)
	}
}

func updateObject(releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) func(oldObj, obj interface{}) {
	return func(oldObj, obj interface{}) {
		handle(obj, releaseGroupSpecName, releaseGroupSpec)
	}
}

func deleteObject(releaseGroupSpecName string, releaseGroupSpec releaseGroupSpec) func(obj interface{}) {
	return func(obj interface{}) {
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
