package driver

import (
	"text/template"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

const (
	// ExperimentSpecPath is the name of the experiment spec file
	ExperimentSpecPath = "experiment.yaml"
	// ExperimentMetricsPathSuffix is the name of the metrics spec file
	ExperimentMetricsPathSuffix = ".metrics.yaml"
	// ExperimentResultPath is the name of the experiment result file
	ExperimentResultPath = "result.yaml"
	// DefaultExperimentGroup is the name of the default experiment chart
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

// MetricsSpecFromBytes reads metrics spec from bytes
func MetricsSpecFromBytes(b []byte) (*template.Template, error) {
	return template.New("metrics-spec").Parse(string(b))
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
