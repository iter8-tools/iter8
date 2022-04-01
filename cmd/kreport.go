package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"

	"github.com/spf13/cobra"
)

const kReportDesc = `
Generate a text or HTML report of a Kubernetes experiment.

	$ iter8 k report # same as iter8 k report -o text

or

	$ iter8 k report -o html > report.html # view with browser
`

// newKReportCmd creates the Kubernetes report command
func newKReportCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewReportOpts(kd)

	cmd := &cobra.Command{
		Use:          "report",
		Short:        "Generate report for Kubernetes experiment",
		Long:         kReportDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.KubeRun(outStream)
		},
	}
	// options specific to k report
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings

	// options shared with report
	addOutputFormatFlag(cmd, &actor.OutputFormat)
	return cmd
}

func init() {
	kCmd.AddCommand(newKReportCmd(kd))
}
