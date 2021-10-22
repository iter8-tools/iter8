package core

import (
	"fmt"
)

// Run an experiment
func (e *Experiment) Run() error {
	var err error
	if e.Result == nil {
		e.Result = &ExperimentResult{}
	}
	if e.Result.StartTime == nil {
		err = e.setStartTime()
		if err != nil {
			return err
		}
	}
	for i, t := range e.Tasks {
		Logger.Info("task index " + fmt.Sprintf("%v", i) + " : " + *e.Spec.Tasks[i].Task + " : " + "started")
		err = t.Run(e)
		if err != nil {
			Logger.Error("task index " + fmt.Sprintf("%v", i) + " : " + *e.Spec.Tasks[i].Task + " : " + "failure")
			e.failExperiment()
			return err
		} else {
			ecerr := e.incrementNumCompletedTasks()
			if ecerr != nil {
				return ecerr
			}
			Logger.Info("task index " + fmt.Sprintf("%v", i) + " : " + *e.Spec.Tasks[i].Task + " : " + "completed")
		}
	}
	return nil
}
