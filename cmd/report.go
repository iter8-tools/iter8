package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"

	"github.com/spf13/cobra"
)

const reportDesc = `
Generate a text or HTML report of an experiment.

	$ iter8 report # same as iter8 report -o text

or

	$ iter8 report -o html > report.html # view with browser
`

// newReportCmd creates the report command
func newReportCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewReportOpts(kd)

	cmd := &cobra.Command{
		Use:          "report",
		Short:        "Generate experiment report",
		Long:         reportDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun(outStream)
		},
	}
	addOutputFormatFlag(cmd, &actor.OutputFormat)
	addRunDirFlag(cmd, &actor.RunDir)
	return cmd
}

// addOutputFormatFlag adds output format flag to the report command
func addOutputFormatFlag(cmd *cobra.Command, outputFormat *string) {
	cmd.Flags().StringVarP(outputFormat, "outputFormat", "o", "text", "text | html")
}

func init() {
	rootCmd.AddCommand(newReportCmd(kd))
}
