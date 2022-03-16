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
This command runs a Kubernetes experiment. It reads an experiment specified in the experiment.yaml file and outputs the result to a Kubernetes secret.

		$ cd /folder/with/experiment.yaml
		$	iter8 k run --namespace {{ .Experiment.Namespace }} --group {{ .Experiment.group }} --revision {{ .Experiment.Revision }}

This command is primarily intended for use within the Iter8 Docker image that is used to execute Kubernetes experiments.
`

func newKRunCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewRunOpts(kd)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "run a Kubernetes experiment",
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
