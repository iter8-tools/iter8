package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

	"github.com/spf13/cobra"
)

// genDesc is the description for the gen command
const genDesc = `
Generate an experiment.yaml file by combining an experiment chart with values.

    $ iter8 gen -c load-test-http --set url=https://httpbin.org/get

This command is intended for development and testing of experiment charts. For production usage, the launch command is recommended.
`

// newGenCmd creates the gen command
func newGenCmd() *cobra.Command {
	actor := ia.NewGenOpts()

	cmd := &cobra.Command{
		Use:          "gen",
		Short:        "Generate experiment.yaml file by combining an experiment chart with values",
		Long:         genDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun()
		},
	}
	addChartsParentDirFlag(cmd, &actor.ChartsParentDir)
	addChartNameFlag(cmd, &actor.ChartName)
	addValueFlags(cmd.Flags(), &actor.Options)
	return cmd
}

// addChartsParentDirFlag to the command
func addChartsParentDirFlag(cmd *cobra.Command, chartsParentDirPtr *string) {
	cmd.Flags().StringVar(chartsParentDirPtr, "chartsParentDir", ".", "directory under which the charts folder is located")
}

// addChartNameFlag to the command
func addChartNameFlag(cmd *cobra.Command, chartNamePtr *string) {
	cmd.Flags().StringVarP(chartNamePtr, "chartName", "c", ia.DefaultChartName, "name of the experiment chart")
}

// initialize with gen command
func init() {
	rootCmd.AddCommand(newGenCmd())
}
