package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	subjectStrSpec = "strSpec"
	weightLabel    = "iter8.tools/weights"
	conditionType  = "condition"
	jsonPathType   = "jsonPath"
)

// subjects by their name
type subjectsByName map[string]*subject

// subjects by object name
type subjectsByObjName map[string]*subject

// subjects by gvr and obj
type subjectsByGVRAndObj map[string]subjectsByObjName

// subjects contain all subjects known to Iter8
type subjects struct {
	mutex sync.RWMutex
	// map each namespace to its subjectsByName
	nsSub map[string]subjectsByName
	// map each namespace to its subjectsByGVRAndObj
	nsObj map[string]subjectsByGVRAndObj
}

var allSubjects = subjects{}

func (s *subjects) getSubFromObj(obj interface{}, gvkrShort string) (*subject, bool) {
	// lock for reading and later unlock
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// get namespace
	u := obj.(*unstructured.Unstructured)
	namespace := u.GetNamespace()

	// attempt to return the subject
	if _, ok1 := allSubjects.nsObj[namespace]; ok1 {
		if _, ok2 := allSubjects.nsObj[namespace][gvkrShort]; ok2 {
			sub, ok3 := allSubjects.nsObj[namespace][gvkrShort][u.GetName()]
			return sub, ok3
		}
	}
	return nil, false
}

func (s *subjects) delete(obj interface{}, config *Config) {
	// lock for writing and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// get namespace and name
	u := obj.(*unstructured.Unstructured)
	namespace := u.GetNamespace()
	name := u.GetName()

	// delete from nsSub
	if m, ok1 := allSubjects.nsSub[namespace]; ok1 {
		delete(m, name)
	}
	// delete from nsObj
	byGVRAndObj := allSubjects.nsObj[namespace]
	for gvkrShort, byGvkr := range byGVRAndObj {
		gvkr, _ := config.KnownGVKRs[gvkrShort]
		if gvkr.matches(u) {
			for objName, _ := range byGvkr {
				if objName == name {
					delete(byGvkr, objName)
				}
			}
		}
	}
}

func (s *subjects) makeAndUpdateWith(obj interface{}) *subject {
	// lock for writing and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// get the subject configmap
	u := obj.(*unstructured.Unstructured)
	cm := &corev1.ConfigMap{}
	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(u.UnstructuredContent(), cm); err != nil {
		e := errors.New("unable to extract subject configmap from object")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return nil
	}

	// validate the configmap
	if err := validateSubjectCM(cm); err != nil {
		return nil
	}

	// make/update subject with uninitialized status
	var sub *subject
	var err error
	if sub, err = extractSubject(cm); err != nil {
		return nil
	}

	// insert into nsSub
	if _, ok := s.nsSub[cm.Namespace]; !ok {
		s.nsSub[cm.Namespace] = make(subjectsByName)
	}
	s.nsSub[cm.Namespace][cm.Name] = sub

	// insert into nsObj
	if _, ok := s.nsObj[cm.Namespace]; !ok {
		s.nsObj[cm.Namespace] = make(subjectsByGVRAndObj)
	}
	for _, v := range sub.Variants {
		for _, r := range v.Resources {
			if _, ok := s.nsObj[cm.Namespace][r.GVKRShort]; !ok {
				s.nsObj[cm.Namespace][r.GVKRShort] = make(subjectsByObjName)
			}
			s.nsObj[cm.Namespace][r.GVKRShort][cm.Name] = sub
		}
	}

	return sub
}

func validateSubjectCM(confMap *corev1.ConfigMap) error {
	if confMap.Immutable == nil || !(*confMap.Immutable) {
		err := errors.New("subject configmap is not immutable")
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

	// initialize weights from variants
	s.weights = make([]uint32, len(s.Variants))
	for i, v := range s.Variants {
		if v.Weight == nil {
			s.weights[i] = 1
		} else {
			s.weights[i] = *v.Weight
		}
	}

	// extract and update weights from label, if present
	if weightStr, ok := confMap.GetObjectMeta().GetLabels()[weightLabel]; ok {
		newWeights := []uint32{}
		if err := yaml.Unmarshal([]byte(weightStr), &newWeights); err == nil {
			if len(newWeights) == len(s.Variants) {
				s.weights = newWeights
			} else {
				log.Logger.Error(fmt.Sprintf("weight array in label has length %d ; expected %d", len(newWeights), len(s.Variants)))
			}
		} else {
			log.Logger.WithStackTrace("error unmarshaling weight label").Error(err)
		}
	}

	// add metadata
	s.ObjectMeta = confMap.ObjectMeta

	return &s, nil
}

func (s *subject) normalizeWeights(config *Config) {
	s.normalizedWeights = make([]uint32, len(s.Variants))
	a := s.getVariantsAvailable(config)
	for i, _ := range s.Variants {
		if a[i] {
			s.normalizedWeights[i] = s.weights[i]
		} else {
			s.normalizedWeights[i] = 0
		}
	}
}

func (s *subject) getVariantsAvailable(config *Config) []bool {
	// initialize all variants for this subject as available
	// if any resource for a variant is unavailable, mark that variant as unavailable
	variantsAvailable := make([]bool, len(s.Variants))
	for i, _ := range variantsAvailable {
		variantsAvailable[i] = true
	}
variantLoop:
	for i, v := range s.Variants {
		for _, r := range v.Resources {
			// get informer for resource, else mark this resource as unavailable
			if _, ok := appInformers[r.GVKRShort]; !ok {
				log.Logger.Error("found resource spec with unknown gvkrShort: ", r.GVKRShort)
				variantsAvailable[i] = false
				continue variantLoop
			}
			var obj runtime.Object
			var err error
			// get resource, else mark this resource as unavailable
			if obj, err = appInformers[r.GVKRShort].Lister().ByNamespace(s.Namespace).Get(r.Name); err != nil {
				log.Logger.Trace("could not get resource: ", r.Name, " with gvkrShort: ", r.GVKRShort)
				variantsAvailable[i] = false
				continue variantLoop
			}
			// check deletionTimestamp
			u := obj.(*unstructured.Unstructured)
			if u.GetDeletionTimestamp() != nil {
				log.Logger.Trace("resource with deletion timestamp: ", r.Name, " with gvkrShort: ", r.GVKRShort)
				variantsAvailable[i] = false
				continue variantLoop
			}
			// check readiness condition using kubectl logic
			// this should implement both status/condition and json path conditions
			if !conditionsSatisfied(u, r.GVKRShort, config) {
				log.Logger.Trace("resource does not satisfy condition: ", r.Name, " with gvkrShort: ", r.GVKRShort)
				variantsAvailable[i] = false
				continue variantLoop
			}
		}
	}
	return variantsAvailable
}

// the rest of this file has functions derived from ...
// https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/wait/wait.go
func conditionsSatisfied(u *unstructured.Unstructured, gvkrShort string, config *Config) bool {
	for _, c := range config.KnownGVKRs[gvkrShort].Conditions {
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

func (s *subject) reconcile(config *Config) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.normalizeWeights(config)

	// perform server side applies
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
					// find GVK
					gvk := obj.GroupVersionKind()
					// map to GVR
					if gvr, err := config.mapGVKToGVR(gvk); err == nil {
						dc := k8sclient.NewKubeClient(nil)
						if _, err := dc.Dynamic().Resource(*gvr).Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, buf.Bytes(), metav1.PatchOptions{
							FieldManager: "iter8-controller",
							Force:        base.BoolPointer(true),
						}); err != nil {
							log.Logger.WithStackTrace("cannot server-side-apply SSA template result").Error(err)
						}
					}
				}
			}
		}
	}
}
