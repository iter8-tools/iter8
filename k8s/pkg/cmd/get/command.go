package get

import (
	"fmt"

	"github.com/iter8-tools/iter8/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

var example = `
# Get list of experiments running in cluster
%[1]s get
`

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get",
		Short:        "Get a list of experiments running in the current context",
		Example:      fmt.Sprintf(example, "iter8"),
		SilenceUsage: true,
	}

	factory, streams := utils.AddGenericCliOptions(cmd, true)

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

	cmd.Flags().StringVar(&o.experiment, "experiment", "", "experiment")
	return cmd
}
