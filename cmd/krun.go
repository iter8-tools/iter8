package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const kRunDesc = `
This command runs a Kubernetes experiment. It reads an experiment specified in the experiment.yaml file and outputs the result to a Kubernetes secret.

		$	iter8 k run

This command is primarily intended for use within the Iter8 Docker image that is used to execute Kubernetes experiments.
`

func newKRunCmd() *cobra.Command {
	actor := ia.NewRunOpts()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "run a Kubernetes experiment",
		Long:  kRunDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := actor.KubeRun()
			if err != nil {
				log.Logger.Error(err)
				return err
			}
			return nil
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	addLaunchFlags(cmd, actor)
	return cmd
}

// cd /iter8exp \
// iter8 k run --namespace {{ .Release.Namespace }} --group {{ .Release.Name }} --revision {{ .Release.Revision }}
