package core

import (
	"errors"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetVersionRecommendedForPromotion from the experiment.
func (e *Experiment) GetVersionRecommendedForPromotion() (string, error) {
	if e == nil {
		return "", errors.New("function GetVersionRecommendedForPromotion() called on nil experiment")
	}
	if e.Status.VersionRecommendedForPromotion == nil {
		return "", errors.New("version recommended for promotion not found in experiment status")
	}
	return *e.Status.VersionRecommendedForPromotion, nil
}

// GetVersionDetail from the experiment for a named version.
func (e *Experiment) GetVersionDetail(versionName string) (*v2alpha2.VersionDetail, error) {
	if e == nil {
		return nil, errors.New("getVersionDetail(...) called on nil experiment")
	}
	if e.Spec.VersionInfo != nil {
		if e.Spec.VersionInfo.Baseline.Name == versionName {
			return &e.Spec.VersionInfo.Baseline, nil
		}
		for i := 0; i < len(e.Spec.VersionInfo.Candidates); i++ {
			if e.Spec.VersionInfo.Candidates[i].Name == versionName {
				return &e.Spec.VersionInfo.Candidates[i], nil
			}
		}
	}
	return nil, errors.New("no version found with name " + versionName)
}

// GetActionSpec gets a named action spec from an experiment.
func (e *Experiment) GetActionSpec(name string) (v2alpha2.Action, error) {
	if e == nil {
		return nil, errors.New("GetActionSpec(...) called on nil experiment")
	}
	if e.Spec.Strategy.Actions == nil {
		return nil, errors.New("nil actions")
	}
	if actionSpec, ok := e.Spec.Strategy.Actions[name]; ok {
		return actionSpec, nil
	}
	return nil, errors.New("action with name " + name + " not found")
}

// UpdateVariable updates a variable within the given VersionDetail. If the variable is already present in the VersionDetail object, the pre-existing value takes precedence and is retained; if not, the new value is inserted.
func UpdateVariable(v *v2alpha2.VersionDetail, name string, value string) error {
	if v == nil {
		return errors.New("nil valued version detail")
	}
	for i := 0; i < len(v.Variables); i++ {
		if v.Variables[i].Name == name {
			log.Info("variable with same name already present in version detail")
			return nil
		}
	}
	v.Variables = append(v.Variables, v2alpha2.NamedValue{
		Name:  name,
		Value: value,
	})
	return nil
}

// FindVariableInVersionDetail scans the variables slice in the given version detail and returns the value of the given variable.
func FindVariableInVersionDetail(v *v2alpha2.VersionDetail, name string) (string, error) {
	if v == nil {
		return "", errors.New("nil valued VersionDetail")
	}
	for i := 0; i < len(v.Variables); i++ {
		if v.Variables[i].Name == name {
			return v.Variables[i].Value, nil
		}
	}
	return "", errors.New("variable not present in VersionDetail")
}

// SetAggregatedBuiltinHists sets the experiment status field corresponding to aggregated built in hists
func (e *Experiment) SetAggregatedBuiltinHists(fortioData v1.JSON) {
	if e.Status.Analysis == nil {
		e.Status.Analysis = &v2alpha2.Analysis{}
	}
	abh := &v2alpha2.AggregatedBuiltinHists{}
	e.Status.Analysis.AggregatedBuiltinHists = abh
	abh.AnalysisMetaData = v2alpha2.AnalysisMetaData{
		Provenance: "Builtin metrics collector",
		Timestamp:  metav1.Now(),
	}
	abh.Data = fortioData
}
