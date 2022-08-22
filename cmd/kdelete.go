package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// kDeleteDesc is the description of the delete cmd
const kDeleteDesc = `
Delete an experiment (group) in Kubernetes.

	iter8 k delete
`

// newKDeleteCmd deletes an experiment group in Kubernetes.
func newKDeleteCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewDeleteOpts(kd)

	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete an experiment (group) in Kubernetes",
		Long:         kDeleteDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.KubeRun()
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group)
	actor.EnvSettings = settings
	return cmd
}

// initialize with the k delete cmd
func init() {
	kCmd.AddCommand(newKDeleteCmd(kd, os.Stdout))
}
