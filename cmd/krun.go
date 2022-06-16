package cmd

import (
	"io"
	"os"

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
	addExperimentGroupFlag(cmd, &actor.Group, true)
	addReuseResult(cmd, &actor.ReuseResult)
	actor.EnvSettings = settings
	cmd.MarkFlagRequired("namespace")
	return cmd
}

// initialize with k run cmd
func init() {
	kCmd.AddCommand(newKRunCmd(kd, os.Stdout))
}
