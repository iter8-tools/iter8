package k8s

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/cmd/options"
)

var example = `
# Generate Kubernetes manifest
iter6 gen k8s
`

func NewCmd() *cobra.Command {
	o := newOptions()

	cmd := &cobra.Command{
		Use:          "k8s",
		Short:        "Generate manifest for running experiment in Kubernetes",
		Example:      example,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.complete(c, args); err != nil {
				return err
			}
			if err := o.validate(c, args); err != nil {
				return err
			}
			if err := o.run(c, args); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.AddCommand(options.NewCmdOptions(os.Stdout))
	return cmd
}
