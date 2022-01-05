package cmd

import (
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

var logsCmd *cobra.Command

func init() {
	// initialize logsCmd
	logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Get logs of an experiment in Kubernetes",
		Example: `
# Get logs of the most recent experiment started in Kubernetes
iter8 k logs

# Get logs of the most recent experiment with app label $APP
iter8 k logs -a $APP

# Get logs of the experiment in Kubernetes with identifier $ID
iter8 k logs --id $ID`,
		RunE: func(c *cobra.Command, args []string) error {
			k8sExperimentOptions.initK8sExperiment(true)
			log.Logger.Infof("logs for experiment: %s\n", *k8sExperimentOptions.id)
			return GetExperimentLogs(k8sExperimentOptions.client, k8sExperimentOptions.namespace, *k8sExperimentOptions.id)
		},
	}

	// initialize options for logsCmd
	logsCmd.Flags().AddFlag(basecli.GetIdFlag())
	logsCmd.Flags().AddFlag(basecli.GetAppFlag())

	// logsCmd is now initialized
	kCmd.AddCommand(logsCmd)
}
