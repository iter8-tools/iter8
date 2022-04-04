package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

	"github.com/spf13/cobra"
)

const genDesc = `
Generate an experiment.yaml file by combining an experiment chart with values.

    $ iter8 gen --sourceDir /path/to/load-test-http --set url=https://httpbin.org/get

This command is intended for development and testing of experiment charts. For production usage, the launch subcommand is recommended.
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

// addChartsParentDirFlag
func addChartsParentDirFlag(cmd *cobra.Command, chartsParentDirPtr *string) {
	cmd.Flags().StringVar(chartsParentDirPtr, "chartsParentDir", ".", "path to experiment chart directory")
}

// addChartNameFlag
func addChartNameFlag(cmd *cobra.Command, chartNamePtr *string) {
	cmd.Flags().StringVarP(chartNamePtr, "chartName", "c", "", "path to experiment chart directory")
	cmd.MarkFlagRequired("chartName")
}

func init() {
	rootCmd.AddCommand(newGenCmd())
}
