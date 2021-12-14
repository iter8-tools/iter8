package cmd

import (
	"github.com/iter8-tools/iter8/base/log"

	"github.com/spf13/cobra"
)

var logsCmd *cobra.Command

func init() {
	// initialize logsCmd
	logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Get logs from experiment",
		Example: `
# Get logs of more recent experiment started in Kubernetes cluster
iter8 k logs

# Get logs of more experiment running in Kubernetes with identifier $EXPERIMENT_ID
iter8 k logs -e $EXPERIMENT_ID`,
		RunE: func(c *cobra.Command, args []string) error {
			k8sExperimentOptions.initK8sExperiment(true)
			log.Logger.Infof("logs for experiment: %s\n", k8sExperimentOptions.experimentId)
			return GetExperimentLogs(k8sExperimentOptions.client, k8sExperimentOptions.namespace, k8sExperimentOptions.experimentId)
		},
	}
	k8sExperimentOptions.addExperimentIdOption(logsCmd.Flags())

	// logsCmd is now initialized
	kCmd.AddCommand(logsCmd)
}
