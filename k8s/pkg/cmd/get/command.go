package get

import (
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var example = `
# Get list of experiments running in cluster
iter8 get
`

func NewCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get",
		Short:        "Get a list of experiments running in the current context",
		Example:      example,
		SilenceUsage: true,
	}

	o := newOptions(streams)

	cmd.RunE = func(c *cobra.Command, args []string) error {
		if err := o.complete(factory, c, args); err != nil {
			return err
		}
		if err := o.validate(c, args); err != nil {
			return err
		}
		if err := o.run(c, args); err != nil {
			return err
		}
		return nil
	}

	cmd.Flags().StringVarP(&o.experimentId, "experiment-id", "e", "", "remote experiment identifier")

	// Prevent default options from being displayed by the help
	utils.HideGenericCliOptions(cmd)

	return cmd
}
