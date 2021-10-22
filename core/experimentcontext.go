package core

import (
	"errors"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Experiment context provides the methods needed to store and retrieve experiment spec and results
type ExperimentContext interface {
	// ReadSpec reads experiment spec
	ReadSpec() (*ExperimentSpec, error)
	// ReadResult reads experiment result
	ReadResult() (*ExperimentResult, error)
	// WriteResult writes experiment result
	WriteResult(*ExperimentResult) error
}

// File context enables file based experiments
type FileContext struct {
	// SpecFile is the full path to the experiment
	SpecFile string
	// ResultFile is the full path to the experiment result
	ResultFile string
}

// ReadSpec from file
func (fc *FileContext) ReadSpec() (*ExperimentSpec, error) {
	yamlFile, err := ioutil.ReadFile(fc.SpecFile)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec file")
	}
	es := ExperimentSpec{}
	err = yaml.Unmarshal(yamlFile, &es)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment spec")
	}
	return &es, err
}

// ReadResult from file
func (fc *FileContext) ReadResult() (*ExperimentResult, error) {
	return nil, nil
}

// WriteResult to file
func (fc *FileContext) WriteResult(r *ExperimentResult) error {
	rBytes, err := yaml.Marshal(r)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to marshal experiment result")
		return errors.New("unable to marshal experiment result")
	}
	err = ioutil.WriteFile(fc.ResultFile, rBytes, 0664)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to write result file")
	}
	return err
}
