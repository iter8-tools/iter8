package cmd

import (
	"fmt"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

var reportCmd *cobra.Command

func init() {
	// initialize reportCmd
	reportCmd = basecli.NewReportCmd()
	var example = `
# Generate a text report for the most recent experiment stared in a Kubernetes cluster
iter8 k report`
	reportCmd.Example = fmt.Sprintf("%s %s\n", reportCmd.Example, example)
	reportCmd.SilenceErrors = true

	reportCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment(true)
		log.Logger.Infof("generating report for experiment: %s\n", k8sExperimentOptions.experimentId)
		return k8sExperimentOptions.experiment.Report(basecli.ReportOptions.OutputFormat)
	}

	k8sExperimentOptions.addExperimentIdOption(reportCmd.Flags())

	// reportCmd is now initialized
	kCmd.AddCommand(reportCmd)
}
