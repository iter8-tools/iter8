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

const (
	experimentSpecPath   = "experiment.yaml"
	experimentResultPath = "result.yaml"
)

// Build an experiment from file
func build(withResult bool) (*Experiment, error) {
	e := &Experiment{
		Experiment: &base.Experiment{},
	}
	var err error
	// read it in
	log.Logger.Trace("build started")
	e.Tasks, err = readSpec()
	if err != nil {
		return nil, err
	}
	e.InitResults()
	if withResult {
		e.Result, err = readResult()
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

// read experiment spec from file
func readSpec() ([]base.TaskSpec, error) {
	yamlFile, err := ioutil.ReadFile(experimentSpecPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return nil, errors.New("unable to read experiment spec")
	}
	e := []base.TaskSpec{}
	err = yaml.Unmarshal(yamlFile, &e)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment spec")
		return nil, err
	}
	return e, err
}

// read experiment result from file
func readResult() (*base.ExperimentResult, error) {
	yamlFile, err := ioutil.ReadFile(experimentResultPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment result")
		return nil, errors.New("unable to read experiment result")
	}
	e := &base.ExperimentResult{}
	err = yaml.Unmarshal(yamlFile, e)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment result")
		return nil, err
	}
	return e, err
}
