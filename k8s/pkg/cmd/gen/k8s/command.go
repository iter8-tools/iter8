package k8s

import (
	"fmt"

	"github.com/spf13/cobra"
)

var example = `
# Generate Kubernetes manifest
%[1]s gen k8s
`

func NewCmd() *cobra.Command {
	o := newOptions()

	cmd := &cobra.Command{
		Use:          "k8s",
		Short:        "Generate manifest for running experiment in Kubernetes",
		Example:      fmt.Sprintf(example, "iter8"),
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

	return cmd
}
