package controllers

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	subjectStrSpec = "strSpec"
)

// subjectsInNamespace maps from the name of a subject to a subject
type subjectsInNamespace map[string]*subject

// allSubjects maps from a namespace to subjectsInNamespace
type allSubjects map[string]subjectsInNamespace

func (as allSubjects) getSubject(obj interface{}, gvkrShort string) (*subject, bool) {
	// get namespace
	u := obj.(*unstructured.Unstructured)
	ns := u.GetNamespace()
	if subjects, ok := as[ns]; !ok {
		log.Logger.Trace("namespace not found in subjectsMap: ", ns)
		return nil, ok
	} else {
		// walk through all subjects and find a match
		for _, s := range subjects {
			for _, v := range s.Variants {
				for _, r := range v.Resources {
					if gvkrShort == r.GVKRShort && u.GetName() == r.Name {
						return s, true
					}
				}
			}
		}
	}
	log.Logger.Trace("object does not belong to any subject. gvkrShort: ", gvkrShort, " name: ", u.GetName())
	return nil, false
}

func (as allSubjects) makeSubject(cmObj interface{}) error {
	// get the subject configmap
	u := cmObj.(*unstructured.Unstructured).UnstructuredContent()
	cm := corev1.ConfigMap{}
	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(u, &cm); err != nil {
		e := errors.New("unable to extract subject configmap from object")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return err
	}

	// validate the configmap
	if err := validateSubjectCM(&cm); err != nil {
		return err
	}

	// make/update subject with uninitialized status
	if s, err := extractSubject(&cm); err != nil {
		return err
	} else {
		// if namespace doesn't exist in subject map, create entry
		if _, ok := as[s.Namespace]; !ok {
			as[s.Namespace] = make(subjectsInNamespace)
		}
		// create/update the subject
		as[s.Namespace][s.Name] = s
		// update status
		s.updateStatus()
	}

	return nil
}

func (as allSubjects) deleteSubject(cmObj interface{}) (err error) {
	// get subject configmap
	var cm *corev1.ConfigMap
	if cm, err = getSubjectCM(cmObj); err != nil {
		return err
	}

	// validate the configmap
	if err = validateSubjectCM(cm); err != nil {
		return err
	}

	// extract and delete subject
	if s, err := extractSubject(cm); err != nil {
		return err
	} else {
		// if namespace doesn't exist in subject map, return
		if _, ok := as[s.Namespace]; !ok {
			return nil
		}
		// delete subject
		delete(as[s.Namespace], s.Name)
	}

	return nil
}

func getSubjectCM(cmObj interface{}) (*corev1.ConfigMap, error) {
	// get the subject configmap
	u := cmObj.(*unstructured.Unstructured).UnstructuredContent()
	cm := &corev1.ConfigMap{}
	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(u, cm); err != nil {
		e := errors.New("unable to extract subject configmap from object")
		log.Logger.WithStackTrace(e.Error()).Error(err)
		return nil, e
	}
	return cm, nil
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
		err := errors.New("unable to find required key in subject configmap data: " + subjectStrSpec)
		log.Logger.Error(err)
		return nil, err
	}

	// get a temporary subject with variants and ssa conf
	s1 := subject{}
	if err := yaml.Unmarshal([]byte(strSpec), &s1); err != nil {
		e := errors.New("cannot unmarshal subject configmap spec")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}
	return &s1, nil

	// extract the final subject
	s2 := subject{}
	s2.TypeMeta = confMap.TypeMeta
	s2.ObjectMeta = confMap.ObjectMeta
	s2.StrSpec = strSpec
	s2.Variants = s1.Variants
	s2.SSATemplates = s1.SSATemplates
	return &s2, nil
}
