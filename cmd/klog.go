package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

const kLogDesc = `
This command fetches the logs for a Kubernetes experiment.

		$	iter8 k log

or 

		$	iter8 k log --group hello
`

func newKLogCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewLogOpts(kd)

	cmd := &cobra.Command{
		Use:   "log",
		Short: "get logs for a Kubernetes experiment",
		Long:  kLogDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if lg, err := actor.KubeRun(); err != nil {
				log.Logger.Error(err)
			} else {
				log.Logger.WithStackTrace(lg).Info("experiment logs from Kubernetes cluster")
			}
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings
	return cmd
}

func init() {
	kCmd.AddCommand(newKLogCmd(kd))
}
