package cmd

import (
	"github.com/iter8-tools/iter8/core"
	task "github.com/iter8-tools/iter8/tasks"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run an experiment",
	Long:  `Run an experiment locally`,
	Run: func(cmd *cobra.Command, args []string) {
		core.Logger.Info("experiment started")
		fc := core.FileContext{
			SpecFile:   specFile,
			ResultFile: resultFile,
		}
		exp := &core.Experiment{
			ExperimentContext: &fc,
			TaskMaker:         &task.TaskMaker{},
		}
		core.Logger.Trace("build started")
		err := exp.Build()
		core.Logger.Trace("build finished")
		if err != nil {
			core.Logger.Error("experiment build failed")
		} else {
			err := exp.Run()
			if err != nil {
				core.Logger.Error("experiment failed")
			} else {
				core.Logger.Info("experiment completed successfully")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&specFile, "spec", "s", "experiment.yaml", "experiment spec yaml file")
	runCmd.Flags().StringVarP(&resultFile, "results", "r", "results.yaml", "experiment results yaml file")
}
