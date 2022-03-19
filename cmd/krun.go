package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

const kRunDesc = `
Run a Kubernetes experiment. This command reads an experiment specified in a secret and writes the result to another secret.

	$ iter8 k run --namespace {{ .Experiment.Namespace }} --group {{ .Experiment.group }} --revision {{ .Experiment.Revision }}

This command is intended for use within the Iter8 Docker image that is used to execute Kubernetes experiments.
`

// newKRunCmd creates the Kubernetes run command
func newKRunCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewRunOpts(kd)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a Kubernetes experiment",
		Long:  kRunDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if err := actor.KubeRun(); err != nil {
				log.Logger.Error(err)
			}
		},
		Hidden: true,
	}
	addExperimentGroupFlag(cmd, &actor.Group, true)
	addExperimentRevisionFlag(cmd, &actor.Revision, true)
	actor.EnvSettings = settings
	cmd.MarkFlagRequired("namespace")
	return cmd
}

func init() {
	kCmd.AddCommand(newKRunCmd(kd, os.Stdout))
}
