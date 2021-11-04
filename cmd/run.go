package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run experiment",
	Long:  `Run the experiment defined in the local file named experiment.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger.Trace("build started")
		exp, err := build(false)
		log.Logger.Trace("build finished")
		if err != nil {
			log.Logger.Error("experiment build failed")
			os.Exit(1)
		} else {
			log.Logger.Info("starting experiment run")
			err := exp.run()
			if err != nil {
				log.Logger.Error("experiment failed")
			} else {
				log.Logger.Info("experiment completed successfully")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

// Run an experiment
func (e *experiment) run() error {
	var err error
	if e.Result == nil {
		e.Result = &base.ExperimentResult{}
	}
	if e.Result.StartTime == nil {
		err = e.setStartTime()
		if err != nil {
			return err
		}
	}
	for i, t := range e.tasks {
		log.Logger.Info("task " + fmt.Sprintf("%v", i) + " : started")
		err = t.Run(e.Experiment)
		if err != nil {
			log.Logger.Error("task " + fmt.Sprintf("%v", i) + " : " + "failure")
			e.failExperiment()
			return err
		} else {
			e.incrementNumCompletedTasks()
			err = write(e)
			if err != nil {
				return err
			}
			log.Logger.Info("task " + fmt.Sprintf("%v", i) + " : " + "completed")
		}
	}
	return nil
}

// write an experiment to a file
func write(r *experiment) error {
	rBytes, err := yaml.Marshal(r)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to marshal experiment")
		return errors.New("unable to marshal experiment")
	}
	err = ioutil.WriteFile(experimentFilePath, rBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment file")
		return err
	}
	return err
}

func (e *experiment) setStartTime() error {
	if e.Result == nil {
		log.Logger.Warn("setStartTime called on an experiment object without results")
		e.Experiment.InitResults()
	}
	return nil
}

func (e *experiment) failExperiment() error {
	if e.Result == nil {
		log.Logger.Warn("failExperiment called on an experiment object without results")
		e.Experiment.InitResults()
	}
	e.Result.Failure = true
	return nil
}

func (e *experiment) incrementNumCompletedTasks() error {
	if e.Result == nil {
		log.Logger.Warn("incrementNumCompletedTasks called on an experiment object without results")
		e.Experiment.InitResults()
	}
	e.Result.NumCompletedTasks++
	return nil
}

/*
// Run the given action.
func (a *Action) Run(ctx context.Context) error {
	for i := 0; i < len(*a); i++ {
		log.Info("------ task starting")
		shouldRun := true
		exp, err := GetExperimentFromContext(ctx)
		if err != nil {
			return err
		}
		// if task has a condition
		if cond := (*a)[i].GetIf(); cond != nil {
			// condition evaluates to false ... then shouldRun is false
			program, err := expr.Compile(*cond, expr.Env(exp), expr.AsBool())
			if err != nil {
				return err
			}

			output, err := expr.Run(program, exp)
			if err != nil {
				return err
			}

			shouldRun = output.(bool)
		}
		if shouldRun {
			err := (*a)[i].Run(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
*/
