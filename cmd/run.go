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
# Run experiment with identifier $EXPERIMENT_ID defined in a Kubernetes secret named "experiment-$EXPERIMENT_ID"
iter8 k run -e $EXPERIMENT_ID`
	runCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment(false)
		return k8sExperimentOptions.experiment.Run(k8sExperimentOptions.expIO)
	}

	k8sExperimentOptions.addExperimentIdOption(runCmd.Flags())

	// runCmd is now initialized
	kCmd.AddCommand(runCmd)
}
