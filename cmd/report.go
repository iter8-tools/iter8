package cmd

// import (
// 	"github.com/iter8-tools/iter8/base/log"
// 	"github.com/iter8-tools/iter8/basecli"

// 	"github.com/spf13/cobra"
// )

// var reportCmd *cobra.Command

// func init() {
// 	// initialize reportCmd
// 	reportCmd = basecli.NewReportCmd()
// 	reportCmd.Example = `
// # View a text report for the most recent experiment started in Kubernetes
// iter8 k report

// # View an html report for the most recent experiment
// iter8 k report -o html

// # View an html report for the most recent experiment with app label $APP
// iter8 k report -o html -a $APP

// # View an html report the experiment with identifier $ID
// iter8 k report -o html --id $ID`
// 	reportCmd.SilenceErrors = true

// 	reportCmd.RunE = func(c *cobra.Command, args []string) error {
// 		k8sExperimentOptions.initK8sExperiment(true)
// 		log.Logger.Infof("generating report for experiment: %s\n", *k8sExperimentOptions.id)
// 		return k8sExperimentOptions.experiment.Report(basecli.ReportOptions.OutputFormat)
// 	}

// 	// initialize options for reportCmd
// 	reportCmd.Flags().AddFlag(getIdFlag())
// 	reportCmd.Flags().AddFlag(getAppFlag())

// 	// reportCmd is now initialized
// 	kCmd.AddCommand(reportCmd)
// }
