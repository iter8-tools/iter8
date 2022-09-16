package cmd

import (
	"io"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// kRunDesc is the description of the k run command
const kRunDesc = `
Run a Kubernetes experiment. This command reads an experiment specified in a secret and writes the result back to the secret.

	$ iter8 k run --namespace {{ .Experiment.Namespace }} --group {{ .Experiment.group }}

This command is intended for use within the Iter8 Docker image that is used to execute Kubernetes experiments.
`

// newKRunCmd creates the Kubernetes run command
func newKRunCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewRunOpts(kd)
	actor.EnvSettings = settings
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "Run a Kubernetes experiment",
		Long:         kRunDesc,
		SilenceUsage: true,
		Hidden:       true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.KubeRun()
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group)
	addReuseResult(cmd, &actor.ReuseResult)
	return cmd
}
