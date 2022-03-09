package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const reportDesc = `
This command generates a text or HTML report of an experiment.

    $ iter8 report

or

    $ iter8 report -o html > report.html # view with browser
`

func newReportCmd() *cobra.Command {
	actor := ia.NewReportOpts()

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate experiment report",
		Long:  reportDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := actor.LocalRun(); err != nil {
				log.Logger.Error(err)
				return err
			}
			return nil
		},
	}
	addReportFlags(cmd, actor)
	addRunFlags(cmd, &actor.RunOpts)
	return cmd
}

func addReportFlags(cmd *cobra.Command, actor *ia.ReportOpts) {
	cmd.Flags().StringVarP(&actor.OutputFormat, "outputFormat", "o", "text", "text | html")
}

func init() {
	rootCmd.AddCommand(newReportCmd())
}
