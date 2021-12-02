package assert

import (
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
)

var example = `
	# assert that the most recent experiment running in the Kubernetes context is complete
	iter8 assert --remote -c completed
`

func NewCmd() *cobra.Command {
	cmd := basecli.AssertCmd
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)

	factory, streams := utils.AddGenericCliOptions(cmd)

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

	cmd.Flags().StringVarP(&o.experimentId, "experiment-id", "e", "", "remote experiment identifier; if not specified, the most recent experiment is used")
	cmd.Flags().BoolVarP(&o.remote, "remote", "r", false, "test assertions on remotely executed experiment")

	return cmd
}
