package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// kLogDesc is the description of the k log cmd
const kLogDesc = `
Fetch logs for a Kubernetes experiment.

	iter8 k log
`

// newKLogCmd creates the Kubernetes log command
func newKLogCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewLogOpts(kd)

	cmd := &cobra.Command{
		Use:          "log",
		Short:        "Fetch logs for a Kubernetes experiment",
		Long:         kLogDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if lg, err := actor.KubeRun(); err != nil {
				return err
			} else {
				log.Logger.WithIndentedTrace(lg).Info("experiment logs from Kubernetes cluster")
				return nil
			}
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group)
	actor.EnvSettings = settings
	return cmd
}

// initialize with k log cmd
func init() {
	kCmd.AddCommand(newKLogCmd(kd))
}
