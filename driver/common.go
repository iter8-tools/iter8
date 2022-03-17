package driver

import (
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

const (
	// ExperimentSpecPath is the name of the experiment spec file
	ExperimentSpecPath = "experiment.yaml"
	// ExperimentSpecPath is the name of the experiment result file
	ExperimentResultPath = "result.yaml"
	// ExperimentSpecPath is the name of the official Iter8 experiment chart repo
	DefaultIter8RepoURL = "https://iter8-tools.github.io/hub"
	// ExperimentSpecPath is the name of the default experiment chart
	DefaultExperimentGroup = "default"
)

// SpecFromBytes reads experiment spec from bytes
func SpecFromBytes(b []byte) (base.ExperimentSpec, error) {
	e := base.ExperimentSpec{}
	err := yaml.Unmarshal(b, &e)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment spec")
		return nil, err
	}
	return e, err
}

// ResultFromBytes reads experiment result from bytes
func ResultFromBytes(b []byte) (*base.ExperimentResult, error) {
	r := &base.ExperimentResult{}
	err := yaml.Unmarshal(b, r)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment result")
		return nil, err
	}
	return r, err
}
