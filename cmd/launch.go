package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

const launchDesc = `
Launch an experiment. 

	$ iter8 launch -c load-test-http --set url=https://httpbin.org/get

To create the experiment.yaml file without actually running it, use the dry option.

	$ iter8 launch -c load-test-http \
	  --set url=https://httpbin.org/get \
	  --dry

By default, the current directory is used to download and unpack the experiment chart. Change this location using the destDir option.

	$ iter8 launch -c load-test-http \
	  --set url=https://httpbin.org/get \
	  --destDir /tmp
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
	addChartParentDirFlag(cmd, &actor.ChartsParentDir)
	addGitFolderFlag(cmd, &actor.GitFolder)
	addChartNameFlag(cmd, &actor.ChartName)
	addValueFlags(cmd.Flags(), &actor.Options)
	addRunDirFlag(cmd, &actor.RunDir)
	addNoDownloadFlag(cmd, &actor.NoDownload)

	return cmd
}

// addDryRunFlag adds dry run flag to the launch command
func addDryRunFlag(cmd *cobra.Command, dryRunPtr *bool) {
	cmd.Flags().BoolVar(dryRunPtr, "dry", false, "simulate an experiment launch")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
}

// addNoDownloadFlag adds noDownload flag to the launch command
func addNoDownloadFlag(cmd *cobra.Command, noDownloadPtr *bool) {
	cmd.Flags().BoolVar(noDownloadPtr, "noDownload", false, "reuse local charts dir; do not download from Git")
	cmd.Flags().Lookup("noDownload").NoOptDefVal = "true"
}

func init() {
	rootCmd.AddCommand(newLaunchCmd(kd))
}
