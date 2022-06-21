package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// launchDesc is the description of the launch command
const launchDesc = `
Launch an experiment. 

	$ iter8 launch --set "tasks={http}" \
		--set http.url=https://httpbin.org/get

Use the dry option to simulate an experiment. This creates the experiment.yaml file.

	$ iter8 launch \
	--set http.url=https://httpbin.org/get \
	--dry

You can use various launch flags to control the following:
	1. Whether Iter8 should download the Iter8 experiment chart from a remote URL or reuse local chart.
	2. The remote URL (example, a GitHub URL) from which the Iter8 experiment chart is downloaded.
	3. The local (parent) directory under which the Iter8 experiment chart is nested.
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
	cmd.Flags().BoolVar(dryRunPtr, "dry", false, "simulate an experiment launch; outputs experiment.yaml file")
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
