package report

import (
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
)

var example = `
	# Generate text report for the most recent experiment running in current Kubernetes context
	iter8 report --remote
`

func NewCmd() *cobra.Command {
	cmd := basecli.ReportCmd
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)

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

	cmd.Flags().StringVarP(&o.experiment, "experiment", "e", "", "remote experiment; if not specified, the most recent experiment is used")
	cmd.Flags().BoolVarP(&o.remote, "remote", "r", false, "report on remotely executed experiment")

	return cmd
}
