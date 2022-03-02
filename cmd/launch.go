package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

var (
	// dry indicates that experiment.yaml should be genereated but not run
	dry bool
)

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch an Iter8 experiment.",
	Long: `
Launch an Iter8 experiment by downloading a chart from an Iter8 experiment chart repo, rendering an experiment.yaml file by combining the chart with values, and running the experiment specified in experiment.yaml.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// removing files and dirs matching chartname and dest dir
		if err := cleanChartArtifacts(destDir, chartName); err != nil {
			os.Exit(1)
		}

		// download
		err := hubCmd.RunE(cmd, args)
		if err != nil {
			os.Exit(1)
		}

		// render
		chartPath = path.Join(destDir, chartName)
		err = genCmd.RunE(cmd, args)
		if err != nil {
			os.Exit(1)
		}
		if dry {
			return nil
		}

		// run
		err = runCmd.RunE(cmd, args)
		if err != nil {
			os.Exit(1)
		}
		return err
	},
}

func init() {
	launchCmd.Flags().AddFlagSet(hubCmd.Flags())
	launchCmd.MarkFlagRequired("chartName")

	launchCmd.Flags().BoolVar(&dry, "dry", false, "render experiment.yaml without running the experiment")
	launchCmd.Flags().Lookup("dry").NoOptDefVal = "true"

	addGenOptions(launchCmd.Flags())

	rootCmd.AddCommand(launchCmd)
}
