package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

const kDeleteDesc = `
Delete an experiment group in Kubernetes.

	$ iter8 k delete -g hello
`

// newKDeleteCmd deletes an experiment group in Kubernetes.
func newKDeleteCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewDeleteOpts(kd)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an experiment group in Kubernetes",
		Long:  kDeleteDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if err := actor.KubeRun(); err != nil {
				log.Logger.Error(err)
			}
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings
	return cmd
}

func init() {
	kCmd.AddCommand(newKDeleteCmd(kd, os.Stdout))
}
