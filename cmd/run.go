package cmd

import (
	"github.com/iter8-tools/iter8/core"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run an experiment",
	Long:  `Run an experiment locally`,
	Run: func(cmd *cobra.Command, args []string) {
		core.Logger.Info("experiment run started")
		fc := core.FileContext{
			SpecFile:   specFile,
			ResultFile: resultFile,
		}
		exp := &core.Experiment{
			ExperimentContext: &fc,
		}
		err := exp.Run()
		if err != nil {
			core.Logger.Error("experiment run failed")
		} else {
			core.Logger.Info("experiment run completed successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&specFile, "spec", "s", "experiment.yaml", "experiment spec yaml file")
	runCmd.Flags().StringVarP(&resultFile, "results", "r", "results.yaml", "experiment results yaml file")
}
