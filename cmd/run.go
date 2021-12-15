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
	runCmd.Example = `
# Run experiment with identifier $ID defined in a Kubernetes secret named "experiment-$ID"
iter8 k run --id $ID`
	runCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment(false)
		return k8sExperimentOptions.experiment.Run(k8sExperimentOptions.expIO)
	}

	k8sExperimentOptions.addIdOption(runCmd.Flags())

	// runCmd is now initialized
	kCmd.AddCommand(runCmd)
}
