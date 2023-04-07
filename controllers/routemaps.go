package controllers

import (
	"github.com/iter8-tools/iter8/base/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// allRoutemaps contains all the routemaps known to the controller
var allRoutemaps = routemaps{
	nsRoutemap: make(map[string]routemapsByName),
}

// getRoutemapFromObj extracts a routemap which contains the given object as a variant resource
// ToDo: this function assumes that there is at most one routemap that contains a source;
// a more general idea would be to return a routemap list instead
func (s *routemaps) getRoutemapFromObj(obj interface{}, gvrShort string) *routemap {
	// lock for reading and later unlock
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// get unstructured object
	u := obj.(*unstructured.Unstructured)

	// attempt to return the routemap
	// ToDo: speed up this quadruple-nested for loop
	for _, rmByName := range s.nsRoutemap {
		for _, rm := range rmByName {
			for _, v := range rm.Variants {
				for _, r := range v.Resources {
					// go through every resource in every variant in every routemap in every namespace
					// check for name match between routemap-resource and unstructured-resource
					// check for gvrShort match between routemap-resource and unstructured-resource
					if r.GVRShort == gvrShort && r.Name == u.GetName() {
						// if routemap specifies the resource namespace
						// check for namespace match between routemap-resource and unstructured-resource
						if r.Namespace == nil || *r.Namespace == u.GetNamespace() {
							return rm
						}
					}
				}
			}
		}
	}
	return nil
}

// delete a routemap from routemaps
func (s *routemaps) delete(cm *corev1.ConfigMap) {
	// lock for writing and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// delete from nsRoutemap first
	if m, ok1 := s.nsRoutemap[cm.Namespace]; ok1 {
		_, ok2 := m[cm.Name]
		if ok2 {
			delete(m, cm.Name)
			if len(m) == 0 {
				log.Logger.Debug("no routemaps in namespace ", cm.Namespace)
				delete(s.nsRoutemap, cm.Namespace)
				log.Logger.Debug("deleted namespace ", cm.Namespace, " from routemaps")
			}
		}
	}
}

// makeAndUpdateWith makes a routemap from a configmap and updates routemaps with it
func (s *routemaps) makeAndUpdateWith(cm *corev1.ConfigMap, config *Config) *routemap {
	// lock for writing and later unlock
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// validate the configmap
	if err := validateRoutemapCM(cm); err != nil {
		return nil
	}

	log.Logger.Trace("routemap cm is valid")

	// make/update routemap with uninitialized status
	var rm *routemap
	var err error
	if rm, err = extractRoutemap(cm, config); err != nil {
		return nil
	}

	// insert into nsRoutemap
	if _, ok := s.nsRoutemap[cm.Namespace]; !ok {
		s.nsRoutemap[cm.Namespace] = make(routemapsByName)
	}
	s.nsRoutemap[cm.Namespace][cm.Name] = rm

	return rm
}
