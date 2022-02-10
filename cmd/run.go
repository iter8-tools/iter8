package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/antonmedv/expr"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

var runCmd *cobra.Command

// NewRunCmd creates a new run command
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run experiment specified in experiment.yaml.",
		Long: `
Run experiment specified in experiment.yaml.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Logger.Trace("build called")
			fio := &FileExpIO{}
			exp, err := Build(false, fio)
			log.Logger.Trace("build finished")
			if err != nil {
				os.Exit(1)
			} else {
				log.Logger.Info("starting experiment run")
				err := exp.Run(fio)
				if err != nil {
					log.Logger.Error("exiting with code 1")
					os.Exit(1)
				} else {
					log.Logger.Info("experiment completed successfully")
				}
			}
			return nil
		},
	}

	return cmd
}

func init() {
	runCmd = NewRunCmd()
	RootCmd.AddCommand(runCmd)
}

// Run an experiment
func (e *Experiment) Run(expio ExpIO) error {
	var err error
	if e.Result == nil {
		e.InitResults()
	}
	for i, t := range e.Tasks {
		log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + " : started")
		shouldRun := true
		// if task has a condition
		if cond := getIf(t); cond != nil {
			// condition evaluates to false ... then shouldRun is false
			program, err := expr.Compile(*cond, expr.Env(e), expr.AsBool())
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("unable to compile if clause")
				return err
			}

			output, err := expr.Run(program, e)
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("unable to run if clause")
				return err
			}

			shouldRun = output.(bool)
		}
		if shouldRun {
			err = t.Run(&e.Experiment)
			if err != nil {
				log.Logger.Error("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + " : " + "failure")
				e.failExperiment()
				return err
			}
			log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + " : " + "completed")
		} else {
			log.Logger.WithStackTrace(fmt.Sprint("false condition: ", *getIf(t))).Info("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + " : " + "skipped")
		}

		err = e.incrementNumCompletedTasks()
		if err != nil {
			return err
		}
		err = expio.WriteResult(e)
		if err != nil {
			return err
		}
	}
	return nil

}

// failExperiment sets the experiment failure status to true
func (e *Experiment) failExperiment() error {
	if e.Result == nil {
		log.Logger.Warn("failExperiment called on an experiment object without results")
		e.Experiment.InitResults()
	}
	e.Result.Failure = true
	return nil
}

// incrementNumCompletedTasks increments the numbere of completed tasks in the experimeent
func (e *Experiment) incrementNumCompletedTasks() error {
	if e.Result == nil {
		log.Logger.Warn("incrementNumCompletedTasks called on an experiment object without results")
		e.Experiment.InitResults()
	}
	e.Result.NumCompletedTasks++
	return nil
}

// getIf returns the condition (if any) which determine
// whether of not if this task needs to run
func getIf(t base.Task) *string {
	var jsonBytes []byte
	var tm base.TaskMeta
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonBytes to TaskMeta
	_ = json.Unmarshal(jsonBytes, &tm)
	return tm.If
}

// getName returns the name of this task
func getName(t base.Task) *string {
	var jsonBytes []byte
	var tm base.TaskMeta
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonBytes to TaskMeta
	_ = json.Unmarshal(jsonBytes, &tm)

	if tm.Task == nil {
		if tm.Run != nil {
			return base.StringPointer(base.RunTaskName)
		}
	} else {
		return tm.Task
	}
	log.Logger.Error("task spec with no name or run value")
	return nil
}
