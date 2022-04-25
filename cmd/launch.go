package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// launchDesc is the description of the launch command
const launchDesc = `
Launch an experiment. 

	$ iter8 launch -c load-test-http --set url=https://httpbin.org/get

Use the dry option to simulate an experiment. This creates experiment.yaml and *.metrics.yaml files.

	$ iter8 launch -c load-test-http \
	  --set url=https://httpbin.org/get \
	  --dry

Launching an experiment requires the Iter8 experiment charts folder. You can use various launch flags to control:
		1. Whether Iter8 should download the experiment charts folder from a remote URL (example, a GitHub URL), or reuse local charts.
		2. The parent directory of the charts folder.
		3. The remote URL (example, a GitHub URL) from which charts are downloaded.
`

// newLaunchCmd creates the launch command
func newLaunchCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:          "launch",
		Short:        "Launch an experiment",
		Long:         launchDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun()
		},
	}
	addDryRunFlag(cmd, &actor.DryRun)
	addChartsParentDirFlag(cmd, &actor.ChartsParentDir)
	addRemoteFolderURLFlag(cmd, &actor.RemoteFolderURL)
	addChartNameFlag(cmd, &actor.ChartName)
	addValueFlags(cmd.Flags(), &actor.Options)
	addRunDirFlag(cmd, &actor.RunDir)
	addNoDownloadFlag(cmd, &actor.NoDownload)

	return cmd
}

// addDryRunFlag adds dry run flag to the launch command
func addDryRunFlag(cmd *cobra.Command, dryRunPtr *bool) {
	cmd.Flags().BoolVar(dryRunPtr, "dry", false, "simulate an experiment launch; outputs experiment.yaml and *.metrics.yaml files")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
}

// addNoDownloadFlag adds noDownload flag to the launch command
func addNoDownloadFlag(cmd *cobra.Command, noDownloadPtr *bool) {
	cmd.Flags().BoolVar(noDownloadPtr, "noDownload", false, "reuse local charts dir; do not download from Git")
	cmd.Flags().Lookup("noDownload").NoOptDefVal = "true"
}

// initialize with the launch cmd
func init() {
	rootCmd.AddCommand(newLaunchCmd(kd))
}
