package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"sync"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	seriyaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/yaml"
)

/* types: begin */

type routemap struct {
	// Todo: prune this down to agra.ObjectMeta instead of metav1.ObjectMeta
	mutex             sync.RWMutex
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Versions          []version                  `json:"versions,omitempty"`
	RoutingTemplates  map[string]routingTemplate `json:"routingTemplates,omitempty"`
	normalizedWeights []uint32
}

type version struct {
	Resources []resource `json:"resources,omitempty"`
	Weight    *uint32    `json:"weight,omitempty"`
}

type resource struct {
	// GVRShort for the resource
	GVRShort string `json:"gvrShort"`
	// Name of the resource
	Name string `json:"name"`
	// Namespace of the resource
	// if this field is nil,
	// controllers will treat namespace of the routemap as the namespace of this resource
	Namespace *string `json:"namespace"`
}

type routingTemplate struct {
	GVRShort string `json:"gvrShort"`
	Template string `json:"template"`
}

// routemaps by their name
type routemapsByName map[string]*routemap

// routemaps contain every routemap known to Iter8
type routemaps struct {
	mutex sync.RWMutex
	// map each namespace to its routemapsByName
	nsRoutemap map[string]routemapsByName
}

/* types: end */

const (
	routemapStrSpec      = "strSpec"
	weightAnnotation     = "iter8.tools/weight"
	defaultVersionWeight = uint32(1)
)

// Weights provide the relative weights for traffic routing between versions
// Intended for use in routemap templates
func (s *routemap) Weights() []uint32 {
	return s.normalizedWeights
}

// normalizeWeights sets the normalized weights for each version of the routemap
//
// the inputs for normalizedWeights include:
// 1. Whether or not versions are available; if a version is unavailable, its derivedWeight is set to zero
// 2. derivedWeights also get inputs from resource annotations
// 3. derivedWeights can also be directly set in the version definition within the routemap
// 4. derivedWeight is defaulted to 1 for each version
//
// normalizedWeights are the same as derivedWeights with one exception.
// When derivedWeights sum up to zero, we set normalizedWeights[0] to 1
// (i.e., version 1 gets non-zero normalizedWeight)
func (s *routemap) normalizeWeights(config *Config) {
	derivedWeights := make([]uint32, len(s.Versions))
	available := s.getAvailableVersions(config)
	// overrides from version resource annotation
	override := s.getWeightOverrides(config)

	for i, v := range s.Versions {
		if available[i] {
			// first, attempt to weight from the version spec
			if v.Weight != nil {
				derivedWeights[i] = *v.Weight
			} else {
				// no version weight specified; default
				derivedWeights[i] = defaultVersionWeight
			}
			// next, attempt to override weight from object annotations
			if override[i] != nil {
				// found weight override for this version
				derivedWeights[i] = *override[i]
			}
		} else {
			// this version is not available; set weight to 0
			derivedWeights[i] = 0
		}
	}

	// if derivedWeights sum up to zero, set normalizedWeight[0] to (the non-zero) default
	total := uint32(0)
	for _, v := range derivedWeights {
		total += v
	}
	if total == 0 {
		// at this point, routemap is validated and guaranteed to have at least one version
		derivedWeights[0] = defaultVersionWeight
	}
	s.normalizedWeights = derivedWeights

}

// getWeightOverrides is looking for weights in the object annotations
// override pointer for a version may be nil, if there are no valid weight annotation for the version
// if a version has multiple resources,
// this function looks for the override in the first resource only
func (s *routemap) getWeightOverrides(config *Config) []*uint32 {
	override := make([]*uint32, len(s.Versions))
	for i, v := range s.Versions {
		if len(v.Resources) > 0 {
			r := v.Resources[0]
			// get informer for resource, else mark this resource as unavailable
			if _, ok := appInformers[r.GVRShort]; !ok {
				log.Logger.Error("found resource spec with unknown gvrShort: ", r.GVRShort)
				continue
			}
			// get resource, else mark this resource as unavailable
			ns := s.Namespace
			if r.Namespace != nil {
				ns = *r.Namespace
			}
			if obj, err1 := appInformers[r.GVRShort].Lister().ByNamespace(ns).Get(r.Name); err1 != nil {
				log.Logger.Trace("could not get resource: ", r.Name, " with gvrShort: ", r.GVRShort)
				log.Logger.Trace(err1)
				continue
			} else {
				u := obj.(*unstructured.Unstructured)
				if weightStr, ok := u.GetAnnotations()[weightAnnotation]; ok {
					if weight64, err2 := strconv.ParseUint(weightStr, 10, 32); err2 == nil {
						weight32 := uint32(weight64)
						override[i] = &weight32
					} else {
						log.Logger.Error("invalid weight annotation")
					}
				} else {
					log.Logger.Trace("no weight annotation for version resource 1")
				}
			}
		}
	}
	return override
}

func (s *routemap) getAvailableVersions(config *Config) []bool {
	// initialize all versions for this routemap as available
	// if any resource for a version is unavailable, mark that version as unavailable
	versionsAvailable := make([]bool, len(s.Versions))
	for i := range versionsAvailable {
		versionsAvailable[i] = true
	}
versionLoop:
	for i, v := range s.Versions {
		for _, r := range v.Resources {
			// get informer for resource, else mark this resource as unavailable
			if _, ok := appInformers[r.GVRShort]; !ok {
				log.Logger.Error("found resource spec with unknown gvrShort: ", r.GVRShort)
				versionsAvailable[i] = false
				continue versionLoop
			}
			var obj runtime.Object
			var err error
			// get resource, else mark this resource as unavailable
			ns := s.Namespace
			if r.Namespace != nil {
				ns = *r.Namespace
			}
			if obj, err = appInformers[r.GVRShort].Lister().ByNamespace(ns).Get(r.Name); err != nil {
				log.Logger.Trace("could not get resource: ", r.Name, " with gvrShort: ", r.GVRShort)
				log.Logger.Trace(err)
				versionsAvailable[i] = false
				continue versionLoop
			}
			// check deletionTimestamp
			u := obj.(*unstructured.Unstructured)
			if u.GetDeletionTimestamp() != nil {
				log.Logger.Trace("resource with deletion timestamp: ", r.Name, " with gvrShort: ", r.GVRShort)
				versionsAvailable[i] = false
				continue versionLoop
			}
			// check readiness condition using kubectl logic
			// this should implement both status/condition and json path conditions
			if !conditionsSatisfied(u, r.GVRShort, config) {
				log.Logger.Trace("resource does not satisfy condition: ", r.Name, " with gvrShort: ", r.GVRShort)
				versionsAvailable[i] = false
				continue versionLoop
			}
		}
	}
	return versionsAvailable
}

// reconcile a routemap
func (s *routemap) reconcile(config *Config, client k8sclient.Interface) {
	// lock for reading and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// normalize version weights
	s.normalizeWeights(config)

	// if leader, compute routing policy and perform server side apply
	if leaderStatus, err := leaderIsMe(); leaderStatus && err == nil {
		// for each routing template specified in the routemap
		for rtName, rt := range s.RoutingTemplates {
			// create a template
			tpl := template.New(rtName)
			var err error
			// parse template string
			// ensure no parse errors
			if tpl, err = tpl.Option("missingkey=zero").Parse(string(rt.Template)); err != nil {
				log.Logger.WithStackTrace("invalid and unparseable routing template").Error(err)
				return
			}
			buf := &bytes.Buffer{}
			// ensure no template execution errors
			if err := tpl.Execute(buf, s); err != nil {
				log.Logger.WithStackTrace("invalid and unexecutable routing template").Error(err)
			} else {
				// ensure non-empty result from template execution
				result := buf.Bytes()
				if len(result) == 0 {
					log.Logger.Debug("template execution did not yield result: ", rtName)
				} else {
					// result should be a YAML manifest serialized as bytes
					// unmarshal result into unstructured.Unstructured object
					obj := &unstructured.Unstructured{}
					var decUnstructured = seriyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
					if _, _, err := decUnstructured.Decode(result, nil, obj); err != nil {
						log.Logger.WithStackTrace(err.Error()).Error("invalid and unmarshalable routing template")
					} else {
						// ensure object has name and kind
						// if obj.GetName() == "" || obj.GetKind() == "" {
						// ensure object has kind
						if obj.GetKind() == "" {
							log.Logger.Error("template execution yielded invalid object")
						} else {
							// ensure resource type for the object is known
							gvrc, ok := config.ResourceTypes[rt.GVRShort]
							if !ok {
								log.Logger.Error("unknown gvr: ", rt.GVRShort)
							} else {
								// at this point we have a known resource we can server-side apply
								gvr := gvrc.GroupVersionResource

								// get its JSON serialization
								jsonBytes, err := obj.MarshalJSON()
								if err != nil {
									log.Logger.Error("error marshaling obj into JSON: ", err)
								} else {
									// Add namespace to the obj or else broadcast event will use default namespace
									obj.SetNamespace(s.Namespace)
									if _, err := client.Patch(gvr, s.Namespace, obj.GetName(), jsonBytes); err != nil {
										log.Logger.WithStackTrace(err.Error()).Error("cannot server-side-apply routing template result")
										log.Logger.Error("unstructured patch obj: ", obj)
										log.Logger.Error("unstructured obj json: ", string(jsonBytes))

										broadcastEvent(obj, corev1.EventTypeWarning, "Failed to apply template", fmt.Sprintf("Failed to apply template specified in Iter8 routemap %s", rtName), client)
									} else {
										log.Logger.Info("performed server side apply for: ", s.Name, "; in namespace: ", s.Namespace)

										broadcastEvent(obj, corev1.EventTypeNormal, "Applied template", fmt.Sprintf("Applied template specified in Iter8 routemap %s", rtName), client)
									}
								}
							}
						}
					}
				}
			}
		}
	} else if err == nil {
		log.Logger.Info("not leader")
	}
}

// conditionsSatisfied checks if conditions specific in the config are satisfied in an object
// this function is derived from:
// https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/wait/wait.go
func conditionsSatisfied(u *unstructured.Unstructured, gvrShort string, config *Config) bool {
	// loop through conditions specified in config for this gvr
	for _, c := range config.ResourceTypes[gvrShort].Conditions {
		log.Logger.Info("found condition: ", c)
		// this condition is currently not satified
		satisfied := false
		conditions, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
		if err != nil || !found || conditions == nil {
			log.Logger.Info("conditions not found in object")
			return false
		}
		// loop through conditions in the status section of the object
		// attempt to match status condition with config condition
		for _, conditionUncast := range conditions {
			condition, ok := conditionUncast.(map[string]interface{})
			if !ok {
				log.Logger.Info("unable to cast condition to map[string]interface{}")
				return false
			}
			name, found, err := unstructured.NestedString(condition, "type")
			if !found || err != nil || !strings.EqualFold(name, c.Name) {
				log.Logger.Trace("condition with no type")
				continue
			}
			status, found, err := unstructured.NestedString(condition, "status")
			if !found || err != nil {
				log.Logger.Trace("condition with no status")
				continue
			}

			// found a match between config condition and status condition
			generation, found, _ := unstructured.NestedInt64(u.Object, "metadata", "generation")
			if found {
				observedGeneration, found := getObservedGeneration(u, condition)
				if found && observedGeneration < generation {
					// condition generation does not equal resource generation
					log.Logger.Trace("condition not satisfied")
					return false
				}
			}

			// check if condition status equals the required value specified in config
			if !strings.EqualFold(status, c.Status) {
				log.Logger.Info("condition not satisfied")
				return false
			}
			satisfied = true
		}
		if !satisfied {
			// this condition is still not satisfied
			return false
		}
	}
	return true
}

// getObservedGeneration attempts to get the observed generation value from a condition field
// this is best effort and assumes api conventions are followed
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
// this function is derived from:
// https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/wait/wait.go
func getObservedGeneration(obj *unstructured.Unstructured, condition map[string]interface{}) (int64, bool) {
	conditionObservedGeneration, found, _ := unstructured.NestedInt64(condition, "observedGeneration")
	if found {
		return conditionObservedGeneration, true
	}
	statusObservedGeneration, found, _ := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
	return statusObservedGeneration, found
}

// validate routemap CM
func validateRoutemapCM(confMap *corev1.ConfigMap) error {
	// routemap CM must be immutable
	if confMap.Immutable == nil || !(*confMap.Immutable) {
		err := errors.New("routemap CM is not immutable")
		log.Logger.Error(err)
		return err
	}
	return nil
}

// validateRoutemap validates a given routemap
func validateRoutemap(s *routemap, config *Config) (*routemap, error) {
	// routemap must have at least one version
	if len(s.Versions) == 0 {
		e := errors.New("routemap must at least one version")
		log.Logger.Error(e)
		return nil, e
	}

	// if !clusterScoped, version resource namespace should be nil or equal routemap namespace
	if !config.ClusterScoped {
		for _, v := range s.Versions {
			for _, r := range v.Resources {
				if r.Namespace != nil && *r.Namespace != s.Namespace {
					e := errors.New("expected version resource namespace to match routemap namespace")
					log.Logger.Error(e)
					return nil, e
				}
			}
		}
	}

	return s, nil
}

// extractRoutemap from a given configmap
// routemap is also validated
func extractRoutemap(confMap *corev1.ConfigMap, config *Config) (*routemap, error) {
	// get strSpec from the configmap
	strSpec, ok := confMap.Data[routemapStrSpec]
	if !ok {
		err := errors.New("unable to find routemap spec key in configmap")
		log.Logger.Error(err)
		return nil, err
	}

	// unmarshal the routemap from strSpec
	s := routemap{}
	if err := yaml.Unmarshal([]byte(strSpec), &s); err != nil {
		e := errors.New("cannot unmarshal routemap CM spec")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	// transfer configmap metadata to routemap
	s.ObjectMeta = confMap.ObjectMeta

	// validate and return
	return validateRoutemap(&s, config)
}
