package run

import (
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
)

var example = `
	# Run an experiment in a Kubernetes cluster
	iter8 gen -o k8s | kubectl apply -f -
`

func NewCmd() *cobra.Command {
	cmd := basecli.RunCmd
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)

	factory, streams := utils.AddGenericCliOptions(cmd, false)

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

	cmd.Flags().StringVarP(&o.experiment, "experiment", "e", "", "experiment; if not specified, the most recently created one is used")
	cmd.Flags().MarkHidden("experiment")
	cmd.Flags().BoolVarP(&o.remote, "remote", "r", false, "use remote stored experiment")
	cmd.Flags().MarkHidden("remote")
	// TODO make global options hidden

	return cmd
}
