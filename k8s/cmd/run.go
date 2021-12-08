package cmd

import (
	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewRunCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := newK8sExperimentOptions(streams)

	cmd := basecli.NewRunCmd()
	// 	cmd.Example = `# run experimebt using Kubernetes secrets instead of files
	// iter8 k run -e experiment-id`
	cmd.Hidden = true
	cmd.SilenceUsage = true
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		// precompute commonly used values derivable from GetOptions
		return o.initK8sExperiment(factory)
		// add any additional precomutation and/or validation here
	}
	cmd.RunE = func(c *cobra.Command, args []string) error {
		return o.experiment.Run(o.expIO)
	}

	AddExperimentIdOption(cmd, o)
	// Add any other options here

	// Prevent default options from being displayed by the help
	HideGenericCliOptions(cmd)

	return cmd
}
