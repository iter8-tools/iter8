package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

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

func newKReportCmd() *cobra.Command {
	actor := ia.NewReportOpts()

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate report for Kubernetes experiment",
		Long:  kReportDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := actor.KubeRun(); err != nil {
				log.Logger.Error(err)
				return err
			}
			return nil
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, true)
	actor.EnvSettings = settings
	cmd.MarkFlagRequired("namespace")
	addReportFlags(cmd, actor)
	return cmd
}

func init() {
	rootCmd.AddCommand(newReportCmd())
}
