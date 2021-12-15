package cmd

import (
	"github.com/spf13/cobra"
)

var k8sExperimentOptions = newK8sExperimentOptions()

var kCmd = &cobra.Command{
	Use:   "k",
	Short: "Work with experiments running in Kubernetes",
	Example: `
To generate a Kubernetes manifest for an experiment in 'experiment.yaml',
and run it in Kubernetes, do:

	iter8 gen k8s | kubectl apply -f -
`,
	// There is no action associated with this command
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	RootCmd.AddCommand(kCmd)
	flags := kCmd.PersistentFlags()
	k8sExperimentOptions.ConfigFlags.AddFlags(flags)
}
