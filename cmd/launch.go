package cmd

import (
	"errors"
	"os"
	"path"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// launchDesc is the description of the launch command
const launchDesc = `
Launch an experiment in the local environment. 

	iter8 launch --set "tasks={http}" \
	--set http.url=https://httpbin.org/get

Use the dry option to simulate an experiment. This creates the experiment.yaml file but does not run the experiment.

	iter8 launch \
	--set http.url=https://httpbin.org/get \
	--dry

The launch command creates the 'charts' subdirectory under the current working directory, downloads the Iter8 experiment chart, and places it under 'charts'. This behavior can be controlled using various launch flags.

This command supports setting values using the same mechanisms as in Helm. Please see  https://helm.sh/docs/chart_template_guide/values_files/ for more detailed descriptions. In particular, this command supports the --set, --set-file, --set-string, and -f (--values) options all of which have the same behavior as in Helm.
`

// newLaunchCmd creates the launch command
func newLaunchCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:          "launch",
		Short:        "Launch an experiment in the local environment",
		Long:         launchDesc,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return noDownloadIsRequired(actor, cmd.Flags())
		},
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

// noDownloadIsRequired make noDownload flag required, if 'charts' folder is found
func noDownloadIsRequired(lOpts *ia.LaunchOpts, flags *pflag.FlagSet) error {
	chartsFolderPath := path.Join(lOpts.ChartsParentDir, ia.ChartsFolderName)
	if _, err := os.Stat(chartsFolderPath); !os.IsNotExist(err) {
		if flags.Changed("noDownload") {
			return nil
		} else {
			return errors.New("'charts' folder found; 'noDownload' flag is required")
		}
	}
	return nil
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
