package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/iter8/core"
	"github.com/iter8-tools/iter8/core/log"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run experiment",
	Long:  `Run the experiment defined in the local file named experiment.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger.Trace("build started")
		exp, err := Build(false)
		log.Logger.Trace("build finished")
		if err != nil {
			log.Logger.Error("experiment build failed")
			os.Exit(1)
		} else {
			log.Logger.Info("starting experiment run")
			err := exp.Run()
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
func (e *Experiment) Run() error {
	var err error
	if e.Result == nil {
		e.Result = &core.ExperimentResult{}
	}
	if e.Result.StartTime == nil {
		err = e.setStartTime()
		if err != nil {
			return err
		}
	}
	for i, t := range e.Spec.Tasks {
		log.Logger.Info("task " + fmt.Sprintf("%v", i) + "started")
		err = t.Run(e.Experiment)
		if err != nil {
			log.Logger.Error("task " + fmt.Sprintf("%v", i) + " : " + "failure")
			e.failExperiment()
			return err
		} else {
			e.incrementNumCompletedTasks()
			err = Write(e)
			if err != nil {
				return err
			}
			log.Logger.Info("task " + fmt.Sprintf("%v", i) + " : " + "completed")
		}
	}
	return nil
}

// Write an experiment to a file
func Write(r *Experiment) error {
	rBytes, err := yaml.Marshal(r)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to marshal experiment")
		return errors.New("unable to marshal experiment")
	}
	err = ioutil.WriteFile(ExperimentFilePath, rBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment file")
		return err
	}
	return err
}

func (e *Experiment) setStartTime() error {
	if e.Result == nil {
		log.Logger.Warn("setStartTime called on an experiment object without results")
		e.Experiment.InitResults()
	}
	e.Result.StartTime = core.TimePointer(time.Now())
	return nil
}

func (e *Experiment) failExperiment() error {
	if e.Result == nil {
		log.Logger.Warn("failExperiment called on an experiment object without results")
		e.Experiment.InitResults()
	}
	e.Result.Failure = true
	return nil
}

func (e *Experiment) incrementNumCompletedTasks() error {
	if e.Result == nil {
		log.Logger.Warn("incrementNumCompletedTasks called on an experiment object without results")
		e.Experiment.InitResults()
	}
	e.Result.NumCompletedTasks++
	return nil
}
