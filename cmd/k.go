package cmd

import (
	"github.com/spf13/cobra"
)

var k8sExperimentOptions = newK8sExperimentOptions()

var kCmd = &cobra.Command{
	Use:   "k",
	Short: "Work with experiments running in a Kubernetes cluster",
	Example: `
To run an experiment defined in 'experiment.yaml':
iter8 gen k8s | kubectl apply -f -

To delete an experiment with identifier $EXPERIMENT_ID:
iter8 gen k8s --set id=$EXPERIMENT_ID | kubectl delete -f -`,
	// There is no action associated with this command
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	RootCmd.AddCommand(kCmd)
	k8sExperimentOptions.ConfigFlags.AddFlags(kCmd.PersistentFlags())
}
