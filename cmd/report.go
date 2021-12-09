package cmd

import (
	"fmt"

	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

var reportCmd *cobra.Command

func init() {
	// initialize reportCmd
	reportCmd = basecli.NewReportCmd()

	var example = `
	# Generate text report for the most recent experiment running in current Kubernetes context
	iter8 k report`
	reportCmd.Example = fmt.Sprintf("%s%s\n", reportCmd.Example, example)

	reportCmd.SilenceErrors = true
	reportCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment()
		return k8sExperimentOptions.experiment.Report(basecli.ReportOptions.OutputFormat)
	}

	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a list of experiments running in the current context",
		Example: `
# Get list of experiments running in cluster
iter8 k get`,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			k8sExperimentOptions.initK8sExperiment()
			return runGetCmd(c, args, k8sExperimentOptions)
		},
	}
	k8sExperimentOptions.addExperimentIdOption(getCmd.Flags())

	// getCmd is now initialized
	kCmd.AddCommand(getCmd)
}
