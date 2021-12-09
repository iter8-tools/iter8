package cmd

import (
	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

var runCmd *cobra.Command

func init() {
	// initialize runCmd
	runCmd = basecli.NewRunCmd()

	runCmd.Hidden = true
	runCmd.SilenceUsage = true
	runCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment(false)
		return k8sExperimentOptions.experiment.Run(k8sExperimentOptions.expIO)
	}

	k8sExperimentOptions.addExperimentIdOption(runCmd.Flags())

	// runCmd is now initialized
	kCmd.AddCommand(runCmd)
}
