package cmd

import (
	"io"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// krunDesc is the description of the k run command
const krunDesc = `
Run a performance test on Kubernetes. This command reads a test specified in a secret and writes the result back to the secret.

	$ iter8 k run --namespace {{ namespace }} --test {{ test name }}

This command is intended for use within the Iter8 Docker image that is used to execute Kubernetes tests.
`

// newKRunCmd creates the Kubernetes run command
func newKRunCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewRunOpts(kd)
	actor.EnvSettings = settings
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "Run a performance test on Kubernetes",
		Long:         krunDesc,
		SilenceUsage: true,
		Hidden:       true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.KubeRun()
		},
	}
	addTestFlag(cmd, &actor.Test)
	return cmd
}
