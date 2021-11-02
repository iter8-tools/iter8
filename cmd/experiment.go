package cmd

import (
	"errors"
	"io/ioutil"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

type experiment struct {
	tasks []base.Task
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

	for _, t := range e.Spec.Tasks {
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
		}

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
