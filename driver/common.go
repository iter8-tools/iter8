package driver

import (
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

const (
	// DefaultExperimentGroup is the name of the default experiment chart
	DefaultExperimentGroup = "default"
)

// ExperimentFromBytes reads experiment from bytes
func ExperimentFromBytes(b []byte) (*base.Experiment, error) {
	e := base.Experiment{}
	err := yaml.Unmarshal(b, &e)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment: ", string(b))
		return nil, err
	}
	return &e, err
}
