package controllers

import (
	"bytes"
	"context"
	"text/template"

	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/base/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

/* types: begin */

type subject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	StrSpec           string            `json:"strSpec,omitempty"`
	Variants          []Variant         `json:"variants,omitempty"`
	SSATemplates      map[string]string `json:"ssaTemplates,omitempty"`
	Status            SubjectStatus     `json:"status"`
}

type Variant struct {
	Resources []Resource `json:"resources,omitempty"`
}

type Resource struct {
	GVKRShort string `json:"gvkrShort"`
	Name      string `json:"name"`
}

type SubjectStatus struct {
	Variants []VariantStatus `json:"variants,omitempty"`
}

type VariantStatus struct {
	Resources []ResourceStatus `json:"resources,omitempty"`
}

type ResourceStatus struct {
	Exists bool `json:"exists,omitempty"`
	Valid  bool `json:"valid,omitempty"`
}

/* types: end */

func (s *subject) updateStatus() {
	s.Status.Variants = make([]VariantStatus, len(s.Variants))
	for i, v := range s.Variants {
		s.Status.Variants[i].Resources = make([]ResourceStatus, len(s.Variants[i].Resources))
		for j, r := range v.Resources {
			rs := &s.Status.Variants[i].Resources[j]
			if inf, ok := appInformers[r.GVKRShort]; !ok {
				log.Logger.Error("found resource with unknown gvkrShort: ", r.GVKRShort, " in subject: ", s.Name, " in namespace: ", s.Namespace)
			} else {
				if obj, err := inf.Lister().ByNamespace(s.Namespace).Get(r.Name); err != nil {
					rs.Exists = false
					log.Logger.Error(err)
				} else {
					rs.Exists = true
					if m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj); err != nil {
						rs.Valid = true
						log.Logger.Error(err)
					} else {
						rs.Valid = true
						u := &unstructured.Unstructured{Object: m}
						u.GetLabels()
						// ToDo: condition check will be incorporated here ...
					}
				}

			}
		}
	}
}

func (s *subject) reconcileSSA(config *Config) {
	// reconstruct subject status
	s.updateStatus()

	// perform server side applies
	for ssaTemplateName, ssaTemplate := range s.SSATemplates {
		t := template.New(ssaTemplateName)
		if tpl, err := t.Parse(string(ssaTemplate)); err != nil {
			log.Logger.Error(err)
		} else {
			buf := &bytes.Buffer{}
			if err := tpl.Execute(buf, s); err != nil {
				log.Logger.Error(err)
			} else {
				// decode YAML manifest into unstructured.Unstructured
				obj := &unstructured.Unstructured{}
				if err := yaml.Unmarshal(buf.Bytes(), obj); err != nil {
					log.Logger.Error(err)
				} else {
					// find GVK
					gvk := obj.GroupVersionKind()
					// map to GVR
					if gvr, err := config.mapGVKToGVR(gvk); err == nil {
						dc := k8sclient.NewKubeClient(nil)
						if _, err := dc.Dynamic().Resource(*gvr).Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, buf.Bytes(), metav1.PatchOptions{
							FieldManager: "iter8-controller",
						}); err != nil {
							log.Logger.Error(err)
						}
					}
				}
			}
		}
	}
}
