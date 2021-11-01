package cmd

import (
	"errors"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/iter8/core"
	"github.com/iter8-tools/iter8/core/log"
)

// Experiment type with build and assert methods
type Experiment struct {
	*core.Experiment
}

const (
	ExperimentFilePath = "experiment.yaml"
)

// Build an experiment from file
func Build(withResult bool) (*Experiment, error) {
	// read it in
	log.Logger.Trace("build called")
	e, err := Read()
	if err != nil {
		return nil, err
	}
	if !withResult {
		e.Result = &core.ExperimentResult{}
	}
	return e, err
}

// Read an experiment from a file
func Read() (*Experiment, error) {
	yamlFile, err := ioutil.ReadFile(ExperimentFilePath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment file")
		return nil, errors.New("unable to read experiment file")
	}
	e := Experiment{}
	err = yaml.Unmarshal(yamlFile, e.Experiment)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment")
		return nil, err
	}
	return &e, err
}
