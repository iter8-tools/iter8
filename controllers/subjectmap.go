package controllers

import (
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var allSubjects = subjectsMap{
	nsSub: make(map[string]subjectsMapByName),
}

func (s *subjectsMap) getSubFromObj(obj interface{}, gvrShort string) *subject {
	// lock for reading and later unlock
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// get namespace
	u := obj.(*unstructured.Unstructured)
	namespace := u.GetNamespace()

	// attempt to return the subject
	if subsByName, ok1 := allSubjects.nsSub[namespace]; ok1 {
		for _, sub := range subsByName {
			for _, v := range sub.Variants {
				for _, r := range v.Resources {
					if r.GVRShort == gvrShort && r.Name == u.GetName() {
						return sub
					}
				}
			}
		}
	}
	return nil
}

func (s *subjectsMap) delete(cm *corev1.ConfigMap, config *Config, client k8sclient.Interface) {
	// lock for writing and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// get namespace and name

	// delete from nsSub first
	if m, ok1 := allSubjects.nsSub[cm.Namespace]; ok1 {
		_, ok2 := m[cm.Name]
		if ok2 {
			delete(m, cm.Name)
			if len(m) == 0 {
				log.Logger.Debug("no subjects in namespace ", cm.Namespace)
				delete(allSubjects.nsSub, cm.Namespace)
				log.Logger.Debug("deleted namespace ", cm.Namespace, " from allSubjects")
			}
		}
	}
}

func (s *subjectsMap) makeAndUpdateWith(cm *corev1.ConfigMap) *subject {
	// lock for writing and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// validate the configmap
	if err := validateSubjectCM(cm); err != nil {
		return nil
	}

	log.Logger.Trace("subject cm is valid")

	// make/update subject with uninitialized status
	var sub *subject
	var err error
	if sub, err = extractSubject(cm); err != nil {
		return nil
	}

	// insert into nsSub
	if _, ok := s.nsSub[cm.Namespace]; !ok {
		s.nsSub[cm.Namespace] = make(subjectsMapByName)
	}
	s.nsSub[cm.Namespace][cm.Name] = sub

	return sub
}
