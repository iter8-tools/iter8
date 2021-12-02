package assert

import (
	"errors"
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var example = `
	# assert that the most recent experiment running in the Kubernetes context is complete
	iter8 assert --remote -c completed
`

func NewCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := newOptions(streams)

	cmd := basecli.AssertCmd
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if err := o.complete(factory, c, args); err != nil {
			return err
		}
		if err := o.validate(c, args); err != nil {
			return err
		}
		if err := o.run(c, args); err != nil {
			return errors.New("experiment build failed")
		}
		return nil
	}

	cmd.Flags().StringVarP(&o.experiment, "experiment", "e", "", "remote experiment; if not specified, the most recent experiment is used")
	cmd.Flags().BoolVarP(&o.remote, "remote", "r", false, "test assertions on remotely executed experiment")

	return cmd
}
