package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// kLaunchDesc is the description of the k launch cmd
const kLaunchDesc = `
Launch an experiment inside a Kubernetes cluster. 

	iter8 k launch --set "tasks={http}" --set http.url=https://httpbin.org/get \
	--set runner=job

Use the dry option to simulate a Kubernetes experiment. This creates the manifest.yaml file, but does not run the experiment, and does not deploy any experiment resource objects in the cluster.

	iter8 k launch \
	--set http.url=https://httpbin.org/get \
	--set runner=job \
	--dry

The launch command creates the 'charts' subdirectory under the current working directory, downloads the Iter8 experiment chart, and places it under 'charts'. This behavior can be controlled using various launch flags.

This command supports setting values using the same mechanisms as in Helm. Please see  https://helm.sh/docs/chart_template_guide/values_files/ for more detailed descriptions. In particular, this command supports the --set, --set-file, --set-string, and -f (--values) options all of which have the same behavior as in Helm.
`

// newKLaunchCmd creates the Kubernetes launch command
func newKLaunchCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:          "launch",
		Short:        "Launch an experiment inside a Kubernetes cluster",
		Long:         kLaunchDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.KubeRun()
		},
	}
	// flags specific to k launch
	addExperimentGroupFlag(cmd, &actor.Group)
	addDryRunForKFlag(cmd, &actor.DryRun)
	actor.EnvSettings = settings

	// flags shared with launch
	addChartsParentDirFlag(cmd, &actor.ChartsParentDir)
	addRemoteFolderURLFlag(cmd, &actor.RemoteFolderURL)
	addChartNameFlag(cmd, &actor.ChartName)
	addValueFlags(cmd.Flags(), &actor.Options)
	addNoDownloadFlag(cmd, &actor.NoDownload)

	return cmd
}

// addDryRunForKFlag adds dry run flag to the k launch command
func addDryRunForKFlag(cmd *cobra.Command, dryRunPtr *bool) {
	cmd.Flags().BoolVar(dryRunPtr, "dry", false, "simulate an experiment launch; outputs manifest.yaml file")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
}

// initialize with the k launch cmd
func init() {
	kCmd.AddCommand(newKLaunchCmd(kd, os.Stdout))
}
