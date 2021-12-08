package cmd

import (
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type ReportOptions struct {
	// options common to all the k8s commands
	K8sExperimentOptions
	// add other options here
}

func NewReportCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := &GetOptions{K8sExperimentOptions: newK8sExperimentOptions(streams)}

	cmd := basecli.NewReportCmd()
	var example = `
# Generate text report for the most recent experiment running in current Kubernetes context
iter8 report --remote`
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)
	cmd.SilenceUsage = true
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		// precompute commonly used values derivable from GetOptions
		return o.initK8sExperiment(factory)
		// add any additional precomutation and/or validation here
	}
	cmd.RunE = func(c *cobra.Command, args []string) error {
		return o.experiment.Report(basecli.ReportOptions.OutputFormat)

	}

	// Add options
	cmd.Flags().StringVarP(&o.experimentId, ExperimentId, ExperimentIdShort, "", ExperimentIdDescription)

	// Prevent default options from being displayed by the help
	HideGenericCliOptions(cmd)

	return cmd
}
