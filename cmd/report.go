package cmd

import (
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

var reportCmd *cobra.Command

func init() {
	// initialize reportCmd
	reportCmd = basecli.NewReportCmd()
	reportCmd.Example = `
# Generate a text report for the most recent experiment started in a Kubernetes cluster
iter8 k report

# Generate an html report for the most recent experiment
iter8 k report -o html

# Generate an html report the experiment with identifier $ID
iter8 k report -o html --id $ID`
	reportCmd.SilenceErrors = true

	reportCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment(true)
		log.Logger.Infof("generating report for experiment: %s\n", k8sExperimentOptions.id)
		return k8sExperimentOptions.experiment.Report(basecli.ReportOptions.OutputFormat)
	}

	k8sExperimentOptions.addIdOption(reportCmd.Flags())

	// reportCmd is now initialized
	kCmd.AddCommand(reportCmd)
}
