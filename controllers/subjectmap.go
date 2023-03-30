package controllers

import (
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// allSubjects contains all the subjects known to the controller
var allSubjects = subjectsMap{
	nsSub: make(map[string]subjectsMapByName),
}

// getSubFromObj extracts a subject which contains the given object as a variant resource
// ToDo: this function assumes that there is at most one subject that contains a source;
// a more general idea would be to return a subject list instead
func (s *subjectsMap) getSubFromObj(obj interface{}, gvrShort string) *subject {
	// lock for reading and later unlock
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// get unstructured object
	u := obj.(*unstructured.Unstructured)

	// attempt to return the subject
	// ToDo: speed up this quadruple-nested for loop
	for _, subsByName := range allSubjects.nsSub {
		for _, sub := range subsByName {
			for _, v := range sub.Variants {
				for _, r := range v.Resources {
					if r.GVRShort == gvrShort && r.Name == u.GetName() {
						if r.Namespace == nil || *r.Namespace == u.GetNamespace() {
							return sub
						}
					}
				}
			}
		}
	}
	return nil
}

// delete a subject from subjectsMap
func (s *subjectsMap) delete(cm *corev1.ConfigMap, config *Config, client k8sclient.Interface) {
	// lock for writing and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

// makeAndUpdateWith makes a subject from a configmap and updates the subject map with it
func (s *subjectsMap) makeAndUpdateWith(cm *corev1.ConfigMap, config *Config) *subject {
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
	if sub, err = extractSubject(cm, config); err != nil {
		return nil
	}

	// insert into nsSub
	if _, ok := s.nsSub[cm.Namespace]; !ok {
		s.nsSub[cm.Namespace] = make(subjectsMapByName)
	}
	s.nsSub[cm.Namespace][cm.Name] = sub

	return sub
}
