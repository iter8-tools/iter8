package cmd

import (
	"errors"
	"io/ioutil"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

type Experiment struct {
	tasks []base.Task
	*base.Experiment
}

type ExpIO interface {
	readSpec() ([]base.TaskSpec, error)
	readResult() (*base.ExperimentResult, error)
	writeResult(r *Experiment) error
}

const (
	experimentSpecPath   = "experiment.yaml"
	experimentResultPath = "result.yaml"
)

// Build an experiment
func Build(withResult bool, expio ExpIO) (*Experiment, error) {
	e := &Experiment{
		Experiment: &base.Experiment{},
	}
	var err error
	// read it in
	log.Logger.Trace("build started")
	e.Tasks, err = expio.readSpec()
	if err != nil {
		return nil, err
	}
	e.InitResults()
	if withResult {
		e.Result, err = expio.readResult()
		if err != nil {
			return nil, err
		}
	}

	for _, t := range e.Tasks {
		if (t.Task == nil || len(*t.Task) == 0) && (t.Run == nil) {
			log.Logger.Error("invalid task found without a task name or a run command")
			return nil, errors.New("invalid task found without a task name or a run command")
		}

		var task base.Task

		// this is a run task
		if t.Run != nil {
			task, err = base.MakeRun(&t)
			e.tasks = append(e.tasks, task)
			if err != nil {
				return nil, err
			}
		} else {
			// this is some other task
			switch *t.Task {
			case base.CollectTaskName:
				task, err = base.MakeCollect(&t)
				e.tasks = append(e.tasks, task)
			case base.AssessTaskName:
				task, err = base.MakeAssess(&t)
				e.tasks = append(e.tasks, task)
			default:
				log.Logger.Error("unknown task: " + *t.Task)
				return nil, errors.New("unknown task: " + *t.Task)
			}

			if err != nil {
				return nil, err
			}
		}
	}
	return e, err
}

//FileExpIO enables reading and writing through files
type FileExpIO struct{}

// SpecFromBytes reads experiment spec from bytes
func SpecFromBytes(b []byte) ([]base.TaskSpec, error) {
	e := []base.TaskSpec{}
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

// read experiment spec from file
func (f *FileExpIO) readSpec() ([]base.TaskSpec, error) {
	b, err := ioutil.ReadFile(experimentSpecPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return nil, errors.New("unable to read experiment spec")
	}
	return SpecFromBytes(b)
}

// read experiment result from file
func (f *FileExpIO) readResult() (*base.ExperimentResult, error) {
	b, err := ioutil.ReadFile(experimentResultPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment result")
		return nil, errors.New("unable to read experiment result")
	}
	return ResultFromBytes(b)
}

// write experiment result to file
func (f *FileExpIO) writeResult(r *Experiment) error {
	rBytes, _ := yaml.Marshal(r.Result)
	err := ioutil.WriteFile(experimentResultPath, rBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}
