package basecli

import (
	"errors"
	"io/ioutil"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

// Experiment type that includes a list of runnable tasks derived from the experiment spec
type Experiment struct {
	base.Experiment
}

// ExpIO enables interacting with experiment spec and result stored externally
type ExpIO interface {
	// ReadSpec reads the experiment spec
	ReadSpec() (base.ExperimentSpec, error)
	// ReadResult reads the experiment result
	ReadResult() (*base.ExperimentResult, error)
	// WriteResult writes the experiment result
	WriteResult(r *Experiment) error
}

const (
	experimentSpecPath   = "experiment.yaml"
	experimentResultPath = "result.yaml"
)

// Build an experiment
func Build(withResult bool, expio ExpIO) (*Experiment, error) {
	e := &Experiment{}
	var err error
	// read it in
	log.Logger.Trace("build started")
	e.Tasks, err = expio.ReadSpec()
	if err != nil {
		return nil, err
	}
	e.InitResults()
	if withResult {
		e.Result, err = expio.ReadResult()
		if err != nil {
			return nil, err
		}
	}
	return e, err
}

//FileExpIO enables reading and writing experiment spec and result files
type FileExpIO struct{}

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

// ReadSpec reads experiment spec from file
func (f *FileExpIO) ReadSpec() (base.ExperimentSpec, error) {
	b, err := ioutil.ReadFile(experimentSpecPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return nil, errors.New("unable to read experiment spec")
	}
	return SpecFromBytes(b)
}

// ReadResult reads experiment result from file
func (f *FileExpIO) ReadResult() (*base.ExperimentResult, error) {
	b, err := ioutil.ReadFile(experimentResultPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment result")
		return nil, errors.New("unable to read experiment result")
	}
	return ResultFromBytes(b)
}

// WriteResult writes experiment result to file
func (f *FileExpIO) WriteResult(r *Experiment) error {
	rBytes, _ := yaml.Marshal(r.Result)
	err := ioutil.WriteFile(experimentResultPath, rBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}

// Completed returns true if the experiment is complete
// if the result stanza is missing, this function returns false
func (exp *Experiment) Completed() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.NumCompletedTasks == len(exp.Tasks) {
				return true
			}
		}
	}
	return false
}

// NoFailure returns true if no task int he experiment has failed
// if the result stanza is missing, this function returns false
func (exp *Experiment) NoFailure() bool {
	if exp != nil {
		if exp.Result != nil {
			if !exp.Result.Failure {
				return true
			}
		}
	}
	return false
}
