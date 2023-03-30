package controllers

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"strconv"
	"strings"
	"sync"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

/* types: begin */

type subject struct {
	// Todo: prune this down to agra.ObjectMeta instead of metav1.ObjectMeta
	mutex             sync.RWMutex
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Variants          []variant `json:"variants,omitempty"`
	// ToDo: PartialKubeResourceTemplates instead of SSAs
	SSAs              map[string]ssa `json:"ssas,omitempty"`
	normalizedWeights []uint32
}

type variant struct {
	Resources []resource `json:"resources,omitempty"`
	Weight    *uint32    `json:"weight,omitempty"`
}

type resource struct {
	GVRShort  string  `json:"gvrShort"`
	Name      string  `json:"name"`
	Namespace *string `json:"namespace"`
}

type ssa struct {
	GVRShort string `json:"gvrShort"`
	Template string `json:"template"`
}

// subjects by their name
type subjectsMapByName map[string]*subject

// subjectsMap contain all subjects known to Iter8
type subjectsMap struct {
	mutex sync.RWMutex
	// map each namespace to its subjectsByName
	nsSub map[string]subjectsMapByName
}

/* types: end */

const (
	subjectStrSpec       = "strSpec"
	weightAnnotation     = "iter8.tools/weight"
	defaultVariantWeight = uint32(1)
)

// Weights provide the relative weights for traffic routing between variants
func (s *subject) Weights() []uint32 {
	return s.normalizedWeights
}

// normalizeWeights sets the normalized weights for each variant of the subject
//
// the inputs for normalizedWeights include:
// 1. Whether or not variants are available; if a version is unavailable, its derivedWeight is set to zero
// 2. derivedWeights also get inputs from resource annotations
// 3. derivedWeights can also be directly set in the variant definition within the subject
// 4. derivedWeight is defaulted to 1 for each variant
//
// normalizedWeights are the same as derivedWeights with one exception.
// When derivedWeights sum up to zero, we set normalizedWeights[0] to 1
// (i.e., variant 1 gets non-zero normalizedWeight)
func (s *subject) normalizeWeights(config *Config) {
	derivedWeights := make([]uint32, len(s.Variants))
	available := s.getAvailableVariants(config)
	// overrides from variant resource annotation
	override := s.getWeightOverrides(config)

	for i, v := range s.Variants {
		if available[i] {
			// first, attempt to weight from the variant spec
			if v.Weight != nil {
				derivedWeights[i] = *v.Weight
			} else {
				// no variant weight specified; default
				derivedWeights[i] = defaultVariantWeight
			}
			// next, attempt to override weight from object annotations
			if override[i] != nil {
				// found weight override for this variant
				derivedWeights[i] = *override[i]
			}
		} else {
			// this variant is not available; set weight to 0
			derivedWeights[i] = 0
		}
	}

	// if derivedWeights sum up to zero, set normalizedWeight[0] to (the non-zero) default
	total := uint32(0)
	for _, v := range derivedWeights {
		total += v
	}
	if total == 0 {
		// at this point, subject is validated and guaranteed to have at least one variant
		derivedWeights[0] = defaultVariantWeight
	}
	s.normalizedWeights = derivedWeights

}

// getWeightOverrides is looking for weights in the object annotations
// override pointer for a variant may be nil, if there are no valid weight annotation for the variant
// if a variant has multiple resources,
// this function looks for the override in the first resource only
func (s *subject) getWeightOverrides(config *Config) []*uint32 {
	override := make([]*uint32, len(s.Variants))
	for i, v := range s.Variants {
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
					log.Logger.Trace("no weight annotation for variant resource 1")
				}
			}
		}
	}
	return override
}

func (s *subject) getAvailableVariants(config *Config) []bool {
	// initialize all variants for this subject as available
	// if any resource for a variant is unavailable, mark that variant as unavailable
	variantsAvailable := make([]bool, len(s.Variants))
	for i := range variantsAvailable {
		variantsAvailable[i] = true
	}
variantLoop:
	for i, v := range s.Variants {
		for _, r := range v.Resources {
			// get informer for resource, else mark this resource as unavailable
			if _, ok := appInformers[r.GVRShort]; !ok {
				log.Logger.Error("found resource spec with unknown gvrShort: ", r.GVRShort)
				variantsAvailable[i] = false
				continue variantLoop
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
				variantsAvailable[i] = false
				continue variantLoop
			}
			// check deletionTimestamp
			u := obj.(*unstructured.Unstructured)
			if u.GetDeletionTimestamp() != nil {
				log.Logger.Trace("resource with deletion timestamp: ", r.Name, " with gvrShort: ", r.GVRShort)
				variantsAvailable[i] = false
				continue variantLoop
			}
			// check readiness condition using kubectl logic
			// this should implement both status/condition and json path conditions
			if !conditionsSatisfied(u, r.GVRShort, config) {
				log.Logger.Trace("resource does not satisfy condition: ", r.Name, " with gvrShort: ", r.GVRShort)
				variantsAvailable[i] = false
				continue variantLoop
			}
		}
	}
	return variantsAvailable
}

// reconcile a subject
func (s *subject) reconcile(config *Config, client k8sclient.Interface) {
	// lock for reading and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// normalize variant weights
	s.normalizeWeights(config)

	// if leader, compute routing policy and perform server side apply
	if leaderStatus, err := leaderIsMe(); leaderStatus && err == nil {
		// for each routing template specified in the subject
		for ssaName, ssa := range s.SSAs {
			// create a template
			tpl := template.New(ssaName)
			var err error
			// parse template string
			// ensure no parse errors
			if tpl, err = tpl.Option("missingkey=zero").Parse(string(ssa.Template)); err != nil {
				log.Logger.WithStackTrace("invalid and unparseable routing template").Error(err)
				return
			}
			buf := &bytes.Buffer{}
			// ensure no template execution errors
			if err := tpl.Execute(buf, s); err != nil {
				log.Logger.WithStackTrace("invalid and unexecutable ssa template").Error(err)
			} else {
				// ensure non-empty result from template execution
				result := buf.Bytes()
				if len(result) == 0 {
					log.Logger.Debug("template execution did not yield result: ", ssaName)
				} else {
					// result should be a YAML manifest serialized as bytes
					// unmarshal result into unstructured.Unstructured object
					obj := &unstructured.Unstructured{}
					if err := yaml.Unmarshal(result, obj); err != nil {
						log.Logger.WithStackTrace("invalid and unmarshalable ssa template").Error(err)
					} else {
						// ensure object has name and kind
						if obj.GetName() == "" || obj.GetKind() == "" {
							log.Logger.Error("template execution yielded invalid object")
						} else {
							// eusure resource type for the object is known
							gvrc, ok := config.ResourceTypes[ssa.GVRShort]
							if !ok {
								log.Logger.Error("unknown gvr: ", ssa.GVRShort)
							} else {
								// at this point we known resource we can server-side apply
								gvr := gvrc.GroupVersionResource

								if _, err := client.Resource(gvr).Namespace(s.Namespace).Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(result), metav1.PatchOptions{
									FieldManager: "iter8-controller",
									Force:        base.BoolPointer(true),
								}); err != nil {
									log.Logger.WithStackTrace("cannot server-side-apply SSA template result: " + "\n" + string(result)).Error(err)
								} else {
									log.Logger.Info("performed server side apply for: ", s.Name, "; in namespace: ", s.Namespace)
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
		conditions, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
		if err != nil || !found {
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

// validate subject CM
func validateSubjectCM(confMap *corev1.ConfigMap) error {
	// subject confMap must be immutable
	if confMap.Immutable == nil || !(*confMap.Immutable) {
		err := errors.New("subject configmap is not immutable")
		log.Logger.Error(err)
		return err
	}
	return nil
}

// validateSubject validates a given sbject
func validateSubject(s *subject, config *Config) (*subject, error) {
	// subject must have at least one variant
	if len(s.Variants) == 0 {
		e := errors.New("subject must at least one variant")
		log.Logger.Error(e)
		return nil, e
	}

	// if !clusterScoped, variant resource namespace should be nil or equal subject namespace
	if !config.ClusterScoped {
		for _, v := range s.Variants {
			for _, r := range v.Resources {
				if r.Namespace != nil && *r.Namespace != s.Namespace {
					e := errors.New("expected variant resource namespace to match subject namespace")
					log.Logger.Error(e)
					return nil, e
				}
			}
		}
	}

	return s, nil
}

// extractSubject from a given configmap
// subject is also validated
func extractSubject(confMap *corev1.ConfigMap, config *Config) (*subject, error) {
	// get strSpec from the configmap
	strSpec, ok := confMap.Data[subjectStrSpec]
	if !ok {
		err := errors.New("unable to find subject spec key in configmap")
		log.Logger.Error(err)
		return nil, err
	}

	// unmarshal the subject from strSpec
	s := subject{}
	if err := yaml.Unmarshal([]byte(strSpec), &s); err != nil {
		e := errors.New("cannot unmarshal subject configmap spec")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	// transfer configmap metadata to subject
	s.ObjectMeta = confMap.ObjectMeta

	// validate and return
	return validateSubject(&s, config)
}
