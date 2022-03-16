package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const kReportDesc = `
This command generates a text or HTML report for a Kubernetes experiment.

    $ iter8 k report

or

    $ iter8 k report -o html > report.html # view with browser

You can optionally specify the group to which the Kubernetes experiment belongs.

		$ iter8 k report -g hello
`

func newKReportCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewReportOpts(kd)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate report for Kubernetes experiment",
		Long:  kReportDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if err := actor.KubeRun(outStream); err != nil {
				log.Logger.Error(err)
			}
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings
	addReportFlags(cmd, actor)
	return cmd
}

func init() {
	kCmd.AddCommand(newKReportCmd(kd))
}
