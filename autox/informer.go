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
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// this label is used in trigger objects
	// label always set to true when present
	// AutoX controller will only check for the existence of this label, not its value
	autoXLabel = "iter8.tools/autox"

	// this label is used in secrets (to allow for ownership over the applications)
	// label is set to the name of a release group spec (releaseGroupSpecName)
	// there is a 1:1 mapping of secrets to release group specs
	autoXGroupLabel = "iter8.tools/autox-group"

	iter8  = "iter8"
	argocd = "argocd"

	autoXAdditionalValues = "autoXAdditionalValues"

	nameLabel      = "app.kubernetes.io/name"
	versionLabel   = "app.kubernetes.io/version"
	managedByLabel = "app.kubernetes.io/managed-by"
	trackLabel     = "iter8.tools/track"

	timeout  = 15 * time.Second
	interval = 1 * time.Second
)

var applicationGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

var applicationValuesPath = []string{"spec", "source", "helm", "values"}

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
	UID  string `json:"uid" yaml:"uid"`
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

// shouldCreateApplication will return true if an application should be created
// an application should be created if there is no preexisting application or
// if the values are different from those from the previous application
func shouldCreateApplication(values map[string]interface{}, releaseName string) bool {
	// get application
	uPApp, _ := k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Get(context.TODO(), releaseName, metav1.GetOptions{}) // *unstructured.Unstructured previous application
	if uPApp != nil {
		log.Logger.Debug(fmt.Sprintf("found previous application \"%s\"", releaseName))

		// check if the previous application is managed by Iter8
		// (if it was previously created by Iter8)
		if manager, ok := uPApp.GetLabels()[managedByLabel]; !ok || manager != iter8 {
			log.Logger.Debug(fmt.Sprintf("previous application is not managed by Iter8 \"%s\"", releaseName))
			return false
		}

		// extract values from previous application
		pValuesString, _, err := unstructured.NestedString(uPApp.UnstructuredContent(), applicationValuesPath...) // pValuesString previous values
		if err != nil {
			log.Logger.Warn(fmt.Sprintf("cannot extract values of previous application \"%s\": %s: %s", releaseName, pValuesString, err))
		}

		var pValues map[string]interface{}
		err = yaml.Unmarshal([]byte(pValuesString), &pValues)
		if err != nil {
			log.Logger.Warn(fmt.Sprintf("cannot parse values of previous application \"%s\": %s: %s", releaseName, pValuesString, err))
		}

		log.Logger.Debug(fmt.Sprintf("previous values: \"%s\"\nnew values: \"%s\"", pValues, values))

		shouldCreateApplication := !reflect.DeepEqual(pValues, values)
		if shouldCreateApplication {
			log.Logger.Debug(fmt.Sprintf("replace previous application \"%s\"", releaseName))
		} else {
			log.Logger.Debug(fmt.Sprintf("do not replace previous application \"%s\"", releaseName))
		}

		return shouldCreateApplication
	}

	// there is no preexisting application, so should create one
	return true
}

func executeApplicationTemplate(applicationTemplate string, values applicationValues) (*unstructured.Unstructured, error) {
	tpl, err := base.CreateTemplate(applicationTemplate)
	if err != nil {
		log.Logger.Error("could not create application template: ", err)
		return nil, err
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, values)
	if err != nil {
		log.Logger.Error("could not execute application template: ", err)
		return nil, err
	}

	jsonBytes, err := yaml.YAMLToJSON(buf.Bytes())
	if err != nil {
		log.Logger.Error(fmt.Sprintf("could not convert YAML to JSON: \"%s\": \"%s\"", buf.String(), err))
		return nil, err
	}

	// decode pending application into unstructured.UnstructuredJSONScheme
	// source: https://github.com/kubernetes/client-go/blob/1ac8d459351e21458fd1041f41e43403eadcbdba/dynamic/simple.go#L186
	uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsonBytes)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("could not decode object into unstructured.UnstructuredJSONScheme: \"%s\": \"%s\"", buf.String(), err))
		return nil, err
	}

	return uncastObj.(*unstructured.Unstructured), nil
}

// applyApplication will apply an application based on a release spec
func applyApplication(releaseName string, releaseGroupSpecName string, releaseSpec releaseSpec, namespace string, additionalValues map[string]interface{}) error {
	// get release group spec secret, based on autoX group label
	// secret is assigned as the owner of the application
	labelSelector := fmt.Sprintf("%s=%s", autoXGroupLabel, releaseGroupSpecName)
	secretList, err := k8sClient.clientset.CoreV1().Secrets(argocd).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Logger.Error("could not list release group spec secrets: ", err)
		return err
	}

	// ensure that only one secret is found
	if secretsLen := len(secretList.Items); secretsLen == 0 {
		err = errors.New("expected release group spec secret with label selector" + labelSelector + "but none were found")
		log.Logger.Error(err)
		return err
	} else if secretsLen > 1 {
		err = errors.New("expected release group spec secret with label selector" + labelSelector + "but more than one were found")
		log.Logger.Error(err)
		return err
	}
	secret := secretList.Items[0]

	values := applicationValues{ // template values
		Name:      releaseName,
		Namespace: namespace,

		Owner: owner{
			Name: secret.Name,
			UID:  string(secret.GetUID()), // assign the release group spec secret as the owner of the application
		},

		Chart: releaseSpec,
	}

	// add additionalValues to the values
	// Argo CD will create a new experiment if it sees that the additionalValues are different from the previous experiment
	// additionalValues will contain the pruned labels from the Kubernetes object
	if values.Chart.Values == nil {
		values.Chart.Values = map[string]interface{}{}
	}
	values.Chart.Values[autoXAdditionalValues] = additionalValues

	// check if the pending application will be different from the previous application, if it exists
	// only create a new application if it will be different (the values will be different)
	if s := shouldCreateApplication(values.Chart.Values, releaseName); s {
		// delete previous application if it exists
		uPApp, _ := k8sClient.dynamicClient.Resource(applicationGVR).Namespace(argocd).Get(context.TODO(), releaseName, metav1.GetOptions{}) // *unstructured.Unstructured previous application
		if uPApp != nil {
			if err1 := deleteApplication(releaseName); err1 != nil {
				log.Logger.Error(fmt.Sprintf("could not delete previous application: \"%s\": \"%s\"", releaseName, err))
			}
		}

		// execute application template
		uApp, err := executeApplicationTemplate(tplStr, values)
		if err != nil {
			return err
		}

		// apply application to the K8s cluster
		log.Logger.Debug(fmt.Sprintf("apply application \"%s\"", releaseName))
		err = retry.OnError(
			wait.Backoff{
				Steps:    int(timeout / interval),
				Cap:      timeout,
				Duration: interval,
				Factor:   1.0,
				Jitter:   0.1,
			},
			func(err error) bool {
				log.Logger.Error(err)
				return true
			}, // retry on all failures
			func() error {
				_, err = k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Create(context.TODO(), uApp, metav1.CreateOptions{})
				return err
			},
		)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("could not create application: \"%s\": \"%s\"", releaseName, err))
			return err
		}
	}

	return nil
}

// deleteApplication deletes an application based on a given release name
func deleteApplication(releaseName string) error {
	log.Logger.Debug(fmt.Sprintf("delete application \"%s\"", releaseName))

	err := k8sClient.dynamic().Resource(applicationGVR).Namespace(argocd).Delete(context.TODO(), releaseName, metav1.DeleteOptions{})
	if err != nil {
		log.Logger.Error(fmt.Sprintf("could not delete application \"%s\": \"%s\"", releaseName, err))
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
			if err1 := applyApplication(releaseName, releaseGroupSpecName, releaseSpec, namespace, additionalValues); err1 != nil {
				err = errors.New("one or more Helm release applications failed")
			}

		case deleteAction:
			// if there is an error, keep going forward in the for loop
			if err1 := deleteApplication(releaseName); err1 != nil {
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
//	nameLabel     = "app.kubernetes.io/name"
//	versionLabel = "app.kubernetes.io/version"
//	trackLabel   = "iter8.tools/track"
func pruneLabels(labels map[string]string) map[string]interface{} {
	prunedLabels := map[string]interface{}{}
	for _, l := range []string{autoXLabel, nameLabel, versionLabel, trackLabel} {
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

	// at this point, we know that we are really handling an event for the trigger object
	// name, namespace, and GVR should all match
	log.Logger.Debug(fmt.Sprintf("handle kubernetes resource object: name: \"%s\", namespace: \"%s\", kind: \"%s\", labels: \"%s\"", u.GetName(), u.GetNamespace(), u.GetKind(), u.GetLabels()))

	// namespace and GVR should already match trigger
	ns := u.GetNamespace()
	// Note: GVR is from the release group spec, not available through the obj
	gvr := getGVR(releaseGroupSpec)

	// get (client) object from cluster
	clientU, _ := k8sClient.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})

	// if (client) object exists
	// delete applications if (client) object does not have autoX label
	// then apply applications if (client) object has autoX label
	if clientU != nil {
		// check if autoX label does not exist
		clientLabels := clientU.GetLabels()
		if !hasAutoXLabel(clientLabels) {
			log.Logger.Debugf("delete applications for release group \"%s\" (no %s label)", releaseGroupSpecName, autoXLabel)

			_ = doChartAction(deleteAction, releaseGroupSpecName, releaseGroupSpec, "", nil)

			// if autoX label does not exist, there is no need to apply applications, so return
			return
		}

		// apply applications for the release group
		clientPrunedLabels := pruneLabels(clientLabels)
		_ = doChartAction(applyAction, releaseGroupSpecName, releaseGroupSpec, ns, clientPrunedLabels)
	} else { // delete applications if (client) object does not exist
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
		releaseGroupSpecName := releaseGroupSpecName
		releaseGroupSpec := releaseGroupSpec

		ns := releaseGroupSpec.Trigger.Namespace
		gvr := getGVR(releaseGroupSpec)

		w.factories[releaseGroupSpecName] = dynamicinformer.NewFilteredDynamicSharedInformerFactory(k8sClient.dynamicClient, 0, ns, nil)

		informer := w.factories[releaseGroupSpecName].ForResource(gvr)
		_, err := informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    addObject(releaseGroupSpecName, releaseGroupSpec),
			UpdateFunc: updateObject(releaseGroupSpecName, releaseGroupSpec),
			DeleteFunc: deleteObject(releaseGroupSpecName, releaseGroupSpec),
		})

		if err != nil {
			log.Logger.Error(fmt.Sprintf("cannot add event handler for namespace \"%s\" and GVR \"%s\": \"%s\"", ns, gvr, err))
		}
	}

	return w
}

func (watcher *iter8Watcher) start(stopChannel chan struct{}) {
	for _, f := range watcher.factories {
		f.Start(stopChannel)
	}
}
