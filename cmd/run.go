package cmd

import (
	"os"

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
		exp := &core.Experiment{
			TaskMaker: &task.TaskMaker{},
		}
		core.Logger.Trace("build started")
		err := exp.Build()
		core.Logger.Trace("build finished")
		if err != nil {
			core.Logger.Error("experiment build failed")
			os.Exit(1)
		} else {
			core.Logger.Info("building the experiment started")
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
}
