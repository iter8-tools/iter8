package basecli

import (
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
		Short: "Run an experiment",
		Long:  "Run an experiment",
		Example: `
	# Run experiment defined in file 'experiment.yaml' and write result to 'result.yaml'
	iter8 run
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Logger.Trace("build called")
			// Replace FileExpIO with ClusterExpIO to work with
			// Spec and Results that might be inside the cluster
			fio := &FileExpIO{}
			exp, err := Build(false, fio)
			log.Logger.Trace("build finished")
			if err != nil {
				os.Exit(1)
			} else {
				log.Logger.Info("starting experiment run")
				err := exp.Run(fio)
				if err != nil {
					return err
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
	for i, t := range e.tasks {
		log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, t.GetName()) + " : started")
		shouldRun := true
		// if task has a condition
		if cond := base.GetIf(t); cond != nil {
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
			err = t.Run(e.Experiment)
			if err != nil {
				log.Logger.Error("task " + fmt.Sprintf("%v: %v", i+1, t.GetName()) + " : " + "failure")
				e.failExperiment()
				return err
			}
			log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, t.GetName()) + " : " + "completed")
		} else {
			log.Logger.WithStackTrace(fmt.Sprint("false condition: ", *base.GetIf(t))).Info("task " + fmt.Sprintf("%v: %v", i+1, t.GetName()) + " : " + "skipped")
		}

		e.incrementNumCompletedTasks()
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
