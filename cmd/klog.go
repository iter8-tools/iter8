package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// klogDesc is the description of the k log cmd
const klogDesc = `
Fetch logs for a Kubernetes experiment.

	iter8 k log
`

// newKLogCmd creates the Kubernetes log command
func newKLogCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewLogOpts(kd)

	cmd := &cobra.Command{
		Use:          "log",
		Short:        "Fetch logs for a Kubernetes experiment",
		Long:         klogDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			var lg string
			var err error
			if lg, err = actor.KubeRun(); err != nil {
				return err
			}
			log.Logger.WithIndentedTrace(lg).Info("experiment logs from Kubernetes cluster")
			return nil
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group)
	actor.EnvSettings = settings
	return cmd
}
