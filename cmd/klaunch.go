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
Launch an experiment in Kubernetes. 

	$ iter8 k launch -c load-test-http --set url=https://httpbin.org/get

To locally render the Kubernetes experiment, use the dry option.

	$ iter8 k launch -c load-test-http \
	  --set url=https://httpbin.org/get \
	  --dry

By default, the current directory is used to download and unpack the experiment chart. Change this location using the destDir option.

	$ iter8 k launch -c load-test-http \
	  --set url=https://httpbin.org/get \
	  --destDir /tmp
`

// newKLaunchCmd creates the Kubernetes launch command
func newKLaunchCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:          "launch",
		Short:        "Launch an experiment in Kubernetes",
		Long:         kLaunchDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.KubeRun()
		},
	}
	// flags specific to k launch
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings

	// flags shared with launch
	addDryRunFlag(cmd, &actor.DryRun)
	addChartsParentDirFlag(cmd, &actor.ChartsParentDir)
	addGitFolderFlag(cmd, &actor.GitFolder)
	addChartNameFlag(cmd, &actor.ChartName)
	addValueFlags(cmd.Flags(), &actor.Options)
	addNoDownloadFlag(cmd, &actor.NoDownload)

	return cmd
}

// initialize with the k launch cmd
func init() {
	kCmd.AddCommand(newKLaunchCmd(kd, os.Stdout))
}
