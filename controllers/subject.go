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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

/* types: begin */

type subject struct {
	// Todo: prune this down to agra.ObjectMeta instead of metav1.ObjectMeta
	mutex             sync.RWMutex
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Variants          []variant      `json:"variants,omitempty"`
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
	subjectStrSpec   = "strSpec"
	weightAnnotation = "iter8.tools/weight"
)

func (s *subject) Weights() []uint32 {
	return s.normalizedWeights
}

func (s *subject) normalizeWeights(config *Config) {
	s.normalizedWeights = make([]uint32, len(s.Variants))
	available := s.getAvailableVariants(config)
	override := s.getWeightOverrides(config)
	for i, v := range s.Variants {
		if available[i] {
			// first, attempt to weight from the variant spec
			if v.Weight != nil {
				s.normalizedWeights[i] = *v.Weight
			} else {
				// no variant weight specified; initialize to 1
				s.normalizedWeights[i] = 1
			}
			// next, attempt to override weight from object annotations
			if override[i] != nil {
				// found weight override for this variant
				s.normalizedWeights[i] = *override[i]
			}
		} else {
			// this variant is not available; set weight to 0
			s.normalizedWeights[i] = 0
		}
	}
}

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
					log.Logger.Trace("no weight annotation for variant resouce 1")
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

func (s *subject) reconcile(config *Config, client k8sclient.Interface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.normalizeWeights(config)

	// if leader, perform server side applies
	if leaderStatus, err := leaderIsMe(); leaderStatus && err == nil {
		for ssaName, ssa := range s.SSAs {
			t := template.New(ssaName)
			if tpl, err := t.Parse(string(ssa.Template)); err != nil {
				log.Logger.WithStackTrace("invalid and unparseable ssa template").Error(err)
				return
			} else {
				buf := &bytes.Buffer{}
				if err := tpl.Execute(buf, s); err != nil {
					log.Logger.WithStackTrace("invalid and unexecutable ssa template").Error(err)
				} else {
					// decode YAML manifest into unstructured.Unstructured
					obj := &unstructured.Unstructured{}
					if err := yaml.Unmarshal(buf.Bytes(), obj); err != nil {
						log.Logger.WithStackTrace("invalid and unmarshalable ssa template").Error(err)
					} else {
						gvrc, ok := config.ResourceTypes[ssa.GVRShort]
						if !ok {
							log.Logger.Error("unknown gvr: ", ssa.GVRShort)
							continue
						}
						gvr := schema.GroupVersionResource{
							Group:    gvrc.Group,
							Version:  gvrc.Version,
							Resource: gvrc.Resource,
						}
						result := buf.String()
						if _, err := client.Resource(gvr).Namespace(s.Namespace).Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(result), metav1.PatchOptions{
							FieldManager: "iter8-controller",
							Force:        base.BoolPointer(true),
						}); err != nil {
							log.Logger.WithStackTrace("cannot server-side-apply SSA template result: " + "\n" + result).Error(err)
						} else {
							log.Logger.Info("performed server side apply for: ", s.Name, "; in namespace: ", s.Namespace)
						}
					}
				}
			}
		}
	} else if err == nil {
		log.Logger.Info("not leader")
	}
}

// the rest of this file has functions derived from ...
// https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/wait/wait.go
func conditionsSatisfied(u *unstructured.Unstructured, gvrShort string, config *Config) bool {
	for _, c := range config.ResourceTypes[gvrShort].Conditions {
		conditions, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
		if err != nil {
			log.Logger.Info("conditions not found in object")
			return false
		}
		if !found {
			log.Logger.Info("conditions not found in object")
			return false
		}
		for _, conditionUncast := range conditions {
			condition := conditionUncast.(map[string]interface{})
			name, found, err := unstructured.NestedString(condition, "type")
			if !found || err != nil || !strings.EqualFold(name, c.Name) {
				log.Logger.Trace("cannot find condition type")
				continue
			}
			status, found, err := unstructured.NestedString(condition, "status")
			if !found || err != nil {
				log.Logger.Trace("cannot find condition status")
				continue
			}
			generation, found, _ := unstructured.NestedInt64(u.Object, "metadata", "generation")
			if found {
				observedGeneration, found := getObservedGeneration(u, condition)
				if found && observedGeneration < generation {
					log.Logger.Info("found condition for earlier generation of resource")
					return false
				}
			}
			if !strings.EqualFold(status, c.Status) {
				log.Logger.Info("status in resource condition does not equal required status")
				return false
			}
		}
	}
	return true
}

func getObservedGeneration(obj *unstructured.Unstructured, condition map[string]interface{}) (int64, bool) {
	conditionObservedGeneration, found, _ := unstructured.NestedInt64(condition, "observedGeneration")
	if found {
		return conditionObservedGeneration, true
	}
	statusObservedGeneration, found, _ := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
	return statusObservedGeneration, found
}

func validateSubjectCM(confMap *corev1.ConfigMap) error {
	if confMap.Immutable == nil || !(*confMap.Immutable) {
		err := errors.New("subject configmap is not immutable")
		log.Logger.Error(err)
		return err
	}
	return nil
}

func extractSubject(confMap *corev1.ConfigMap) (*subject, error) {
	// get strSpec
	strSpec, ok := confMap.Data[subjectStrSpec]
	if !ok {
		err := errors.New("unable to find subject spec key in configmap")
		log.Logger.Error(err)
		return nil, err
	}

	// unmarshal the subject
	s := subject{}
	if err := yaml.Unmarshal([]byte(strSpec), &s); err != nil {
		e := errors.New("cannot unmarshal subject configmap spec")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	// add metadata
	s.ObjectMeta = confMap.ObjectMeta

	return &s, nil
}
