package cmd

import (
	"errors"
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type AssertOptions struct {
	// options common to all the k8s commands
	K8sExperimentOptions
	// add other options here
}

// run runs the command
func (o *AssertOptions) run(cmd *cobra.Command, args []string) (err error) {
	allGood, err := o.experiment.Assert(basecli.AssertOptions.Conds, basecli.AssertOptions.Timeout)
	if err != nil || !allGood {
		return err
	}

	if !allGood {
		return errors.New("")
	}

	return nil
}

func NewAssertCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := &AssertOptions{K8sExperimentOptions: newK8sExperimentOptions(streams)}

	cmd := basecli.NewAssertCmd()
	var example = `
# assert that the most recent experiment running in the Kubernetes context is complete
iter8 k assert -c completed`
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)
	cmd.SilenceUsage = true
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		// precompute commonly used values derivable from GetOptions
		return o.initK8sExperiment(factory)
		// add any additional precomutation and/or validation here
	}
	cmd.RunE = func(c *cobra.Command, args []string) error {
		return o.run(c, args)
	}

	// Add options
	cmd.Flags().StringVarP(&o.experimentId, ExperimentId, ExperimentIdShort, "", ExperimentIdDescription)

	// Prevent default options from being displayed by the help
	HideGenericCliOptions(cmd)

	return cmd
}
