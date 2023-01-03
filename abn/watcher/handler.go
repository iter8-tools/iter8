package watcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

const (
	versionLabel   = "app.kubernetes.io/version"
	iter8Finalizer = "iter8.tools/finalizer"
)

var (
	applicationNameRE = regexp.MustCompile(`-candidate-[123456789]\d*$`)
)

// handle constructs the application object from the objects currently in the cluster
func handle(action string, obj *unstructured.Unstructured, config serviceConfig, informerFactories map[string]dynamicinformer.DynamicSharedInformerFactory) {
	log.Logger.Tracef("handle %s called", action)
	defer log.Logger.Trace("handle completed")

	// if action == "DELETE" {
	// 	log.Logger.Fatalf("%+v", obj)
	// }
	// if obj.GetDeletionTimestamp() != nil {
	// 	log.Logger.Fatal("delete timestamp")
	// }

	// get object from cluster (even through we have an unstructured.Unstructured, it is really only the metadata; to get the full object we need to fetch it from the cluster)
	obj, err := getUnstructuredObject(obj)
	if err != nil {
		log.Logger.Debug("unable to fetch object from cluster")
		return
	}

	// add finalizer IF object not being deleted AND finalizer not already there
	if obj.GetDeletionTimestamp() == nil && !containsString(obj.GetFinalizers(), iter8Finalizer) {
		obj.SetFinalizers(append(obj.GetFinalizers(), iter8Finalizer))
		log.Logger.Debug("adding Iter8 finalizer")
		_, err := updateUnstructuredObject(obj)
		if err != nil {
			log.Logger.Warn("unable to add finalizer: ", err.Error())
		}
		return // UPDATE action will trigger handle() to do remaining work
	}

	application := getApplicationNameFromObjectName(obj.GetName())
	namespace := obj.GetNamespace()
	version, _ := getVersion(obj)
	log.Logger.Tracef("handle called for %s/%s (%s))", namespace, application, version)

	// get application configuration
	appConfig := getApplicationConfig(namespace, application, config)
	if appConfig == nil {
		// we found a resource that is not part of an a/b/n test; ignore the object
		log.Logger.Debugf("object for application %s/%s has no a/b/n configuration", namespace, application)
		return
	}

	// get the objects related to the application (using the appConfig as a guide)
	applicationObjs := getApplicationObjects(namespace, application, *appConfig, informerFactories)
	log.Logger.Debugf("identified %d related objects", len(applicationObjs))

	// update the application object by updating the mapping of track to version
	//   get the current application
	a, _ := abnapp.Applications.Read(namespace + "/" + application)

	abnapp.Applications.Lock(namespace + "/" + application)
	defer abnapp.Applications.Unlock(namespace + "/" + application)

	//   clear the current mapping
	a.ClearTracks()

	//   for each track, find the version (from cluster objects) and update the mapping
	for _, track := range trackNames(application, *appConfig) {
		log.Logger.Trace("updateApplication for track ", track)
		version, ok := isTrackReady(track, applicationObjs[track], len(appConfig.Resources))
		log.Logger.Trace("updateApplication for track ", track, " found version ", version, " ", ok)
		if ok {
			a.GetVersion(version, true)
			a.Tracks[track] = version
		}
	}

	log.Logger.Debugf("updated application track map: %s/%s --> %v", namespace, application, a.Tracks)

	if obj.GetDeletionTimestamp() != nil && containsString(obj.GetFinalizers(), iter8Finalizer) {
		// if object is being deleted remove the Iter8 finalizer
		// do here (at end) after updating the ApplicationsMap
		log.Logger.Debug("removing Iter8 finalizer")
		obj.SetFinalizers(removeIter8Finalizer(obj.GetFinalizers()))
		_, err := updateUnstructuredObject(obj)
		if err != nil {
			log.Logger.Warn("unable to remove finalizer: ", err.Error())
		}
	}
}

func removeIter8Finalizer(finalizers []string) []string {
	for i, v := range finalizers {
		if v == iter8Finalizer {
			return append(finalizers[:i], finalizers[i+1:]...)
		}
	}
	return finalizers
}

// getApplicationFromObjectName converts the name of the object to the application name
// it converts a name of the form: app[-candidate-index] to a name of the form: app
func getApplicationNameFromObjectName(name string) string {
	locations := applicationNameRE.FindStringIndex(name)
	if len(locations) == 0 {
		return name
	} else {
		return name[:locations[0]]
	}
}

// getVersion gets application version from VERSION_LABEL label on an object
func getVersion(obj *unstructured.Unstructured) (string, bool) {
	labels := obj.GetLabels()
	v, ok := labels[versionLabel]
	return v, ok
}

func isObjectReady(obj *unstructured.Unstructured, gvr schema.GroupVersionResource, condition *string) bool {
	// no condition to check; is ready
	if condition == nil {
		return true
	}

	// get ConditionStatus
	cs, err := getConditionStatus(obj, *condition)
	if err != nil {
		log.Logger.Error("unable to get status: ", err.Error())
		return false
	}
	if strings.EqualFold(*cs, string(corev1.ConditionTrue)) {
		return true
	}

	// condition not True
	return false
}

// TODO rewrite using NestedStringMap
func getConditionStatus(obj *unstructured.Unstructured, conditionType string) (*string, error) {

	if obj == nil {
		return nil, errors.New("no object")
	}

	obj.GetNamespace()

	resultJSON, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	resultObj := make(map[string]interface{})
	err = json.Unmarshal(resultJSON, &resultObj)
	if err != nil {
		return nil, err
	}

	// get object status
	objStatusInterface, ok := resultObj["status"]
	if !ok {
		return nil, errors.New("object does not contain a status")
	}
	objStatus := objStatusInterface.(map[string]interface{})

	conditionsInterface, ok := objStatus["conditions"]
	if !ok {
		return nil, errors.New("object status does not contain conditions")
	}
	conditions := conditionsInterface.([]interface{})
	for _, conditionInterface := range conditions {
		condition := conditionInterface.(map[string]interface{})
		cTypeInterface, ok := condition["type"]
		if !ok {
			return nil, errors.New("condition does not have a type")
		}
		cType := cTypeInterface.(string)
		if strings.EqualFold(cType, conditionType) {
			conditionStatusInterface, ok := condition["status"]
			if !ok {
				return nil, fmt.Errorf("condition %s does not have a value", cType)
			}
			conditionStatus := conditionStatusInterface.(string)
			return &conditionStatus, nil
		}
	}
	return nil, errors.New("expected condition not found")
}

// trackObject is information about an object found to correspond to a track
type trackObject struct {
	gvr       schema.GroupVersionResource
	condition *string
	object    *unstructured.Unstructured
}

// getApplicationObjects identifies a list of objects related to application based on the name
func getApplicationObjects(namespace string, application string, appConfig appDetails, informerFactories map[string]dynamicinformer.DynamicSharedInformerFactory) map[string][]trackObject {

	// initialize
	var trackToObjectList map[string][]trackObject = map[string][]trackObject{}
	tracks := trackNames(application, appConfig)
	for _, track := range tracks {
		trackToObjectList[track] = []trackObject{}
	}

	// get objects by resource type
	for _, r := range appConfig.Resources {
		lister := informerFactories[namespace].ForResource(r.GroupVersionResource).Lister()
		objs, err := lister.List(labels.NewSelector())
		if err != nil {
			// no such objects; can happen if not deployed
			continue
		}
		// reduce to only those that match expectedObjectNames
		// all objects are of the same type but for different tracks
		for _, obj := range objs {
			ao := obj.(*unstructured.Unstructured)
			if ao.GetDeletionTimestamp() != nil {
				// if being deleted, ignore
				continue
			}
			name := ao.GetName()
			_, ok := trackToObjectList[name]
			if !ok {
				// object is not associated with a known track; ignore
				continue
			}
			// create trackObject and add to list of objects for track
			trackToObjectList[name] = append(trackToObjectList[name], trackObject{gvr: r.GroupVersionResource, condition: r.Condition, object: ao})
		}
	}

	return trackToObjectList
}

// isTrackReady checks that all expected objects for the track exist, that the version is defined (consistently) and that the objects are ready
func isTrackReady(track string, trackObjects []trackObject, expectedNumberTrackObjects int) (string, bool) {

	// all objects exist
	if len(trackObjects) != expectedNumberTrackObjects {
		log.Logger.Debugf("expected %d objects; found %d (track: %s)", expectedNumberTrackObjects, len(trackObjects), track)
		return "", false
	}

	// single version identified on at least one object
	var version string = ""
	for _, to := range trackObjects {
		v, ok := getVersion(to.object)
		if ok {
			if version != "" && version != v {
				// different versions on resources of the same track
				log.Logger.Debugf("inconsistent value for label %s (track: %s)", versionLabel, track)
				return "", false
			}
			version = v
		}
	}
	if version == "" {
		log.Logger.Debugf("no value for label %s found (track: %s)", versionLabel, track)
		return "", false
	}

	// all objects are ready
	for _, to := range trackObjects {
		if !isObjectReady(to.object, to.gvr, to.condition) {
			log.Logger.Debugf("no object found of type %v (track: %s)", to.gvr, track)
			return "", false
		}
	}

	return version, true
}

func updateUnstructuredObject(uObj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr, err := k8sclient.Client.GVR(uObj)
	if err != nil {
		return nil, err
	}

	updatedObj, err := k8sclient.Client.Dynamic().
		Resource(*gvr).Namespace(uObj.GetNamespace()).
		Update(
			context.TODO(),
			uObj,
			metav1.UpdateOptions{},
		)

	return updatedObj, err
}

func getUnstructuredObject(uObj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr, err := k8sclient.Client.GVR(uObj)
	if err != nil {
		return nil, err
	}

	obj, err := k8sclient.Client.Dynamic().
		Resource(*gvr).Namespace(uObj.GetNamespace()).
		Get(
			context.TODO(),
			uObj.GetName(),
			metav1.GetOptions{},
		)

	return obj, err
}
