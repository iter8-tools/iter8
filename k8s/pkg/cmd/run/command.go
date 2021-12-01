package run

import (
	"errors"
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var example = `
	# Run an experiment in a Kubernetes cluster
	iter8 gen -o k8s | kubectl apply -f -
`

func NewCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := newOptions(streams)
	cmd := basecli.RunCmd
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

	cmd.Flags().StringVarP(&o.experiment, "experiment", "e", "", "experiment; if not specified, the most recently created one is used")
	cmd.Flags().MarkHidden("experiment")
	cmd.Flags().BoolVarP(&o.remote, "remote", "r", false, "use remote stored experiment")
	cmd.Flags().MarkHidden("remote")
	// TODO make global options hidden

	return cmd
}
