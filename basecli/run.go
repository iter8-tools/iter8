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

// Dry indicates that run should be a dry run
var Dry bool

// NewRunCmd creates a new run command
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Render `experiment.yaml` and run the experiment.",
		Long: `
Render the file named "experiment.yaml" by combining an experiment chart with values, and run the experiment. This command is intended to be executed from the root of an Iter8 experiment chart. Values may be specified and are processed in the same manner as they are for Helm charts.`,
		Example: `
	# Render experiment.yaml and run the experiment
	iter8 run --set url=https://example.com \
	--set SLOs.error-rate=0 \
	--set SLOs.mean-latency=50 \
	--set SLOs.p90=100 \
	--set SLOs.p'97\.5'=200
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// generate experiment here ...
			err := expCmd.RunE(nil, nil)
			if err != nil {
				os.Exit(1)
			}
			if Dry {
				return nil
			}
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

	cmd.Flags().BoolVar(&Dry, "dry", false, "render experiment.yaml without running the experiment")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
	addGenOptions(cmd.Flags())
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
		log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, *base.GetName(t)) + " : started")
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
			err = t.Run(&e.Experiment)
			if err != nil {
				log.Logger.Error("task " + fmt.Sprintf("%v: %v", i+1, *base.GetName(t)) + " : " + "failure")
				e.failExperiment()
				return err
			}
			log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, *base.GetName(t)) + " : " + "completed")
		} else {
			log.Logger.WithStackTrace(fmt.Sprint("false condition: ", *base.GetIf(t))).Info("task " + fmt.Sprintf("%v: %v", i+1, *base.GetName(t)) + " : " + "skipped")
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
