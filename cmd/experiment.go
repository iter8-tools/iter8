package cmd

import (
	"errors"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

type experiment struct {
	*base.Experiment
}

const (
	experimentFilePath = "experiment.yaml"
)

// Build an experiment from file
func build(withResult bool) (*experiment, error) {
	// read it in
	log.Logger.Trace("build called")
	e, err := read()
	if err != nil {
		return nil, err
	}
	if !withResult {
		e.Result = &base.ExperimentResult{}
	}
	return e, err
}

// read an experiment from a file
func read() (*experiment, error) {
	yamlFile, err := ioutil.ReadFile(experimentFilePath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment file")
		return nil, errors.New("unable to read experiment file")
	}
	e := experiment{
		Experiment: &base.Experiment{},
	}
	err = yaml.Unmarshal(yamlFile, e.Experiment)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment")
		return nil, err
	}
	return &e, err
}
