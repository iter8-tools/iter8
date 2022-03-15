package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

const launchDesc = `
This command launches an Iter8 experiment. 

		$ iter8 launch -c load-test-http --set url=https://httpbin.org/get

To create the experiment.yaml file without running the experiment, use the dry option.

$	iter8 launch -c load-test-http \
	--set url=https://httpbin.org/get \
	--dry

By default, the current directory is used to download and unpack the experiment chart, and run the experiment. Control this using the destDir option.

	$	iter8 launch -c load-test-http \
		--set url=https://httpbin.org/get \
		--destDir /tmp
	
By default, the launch command downloads charts from the official Iter8 chart repo. It is also possible to use third party (helm) repos to host Iter8 experiment charts.

		$	iter8 launch -c load-test-http \
			--repoURL https://great.expectations.pip \
			--set url=https://httpbin.org/get
`

func newLaunchCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch an experiment",
		Long:  launchDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if err := actor.Init(); err != nil {
				log.Logger.Error(err)
				os.Exit(1)
			}
			if err := actor.LocalRun(); err != nil {
				log.Logger.Error(err)
				os.Exit(1)
			}
		},
	}
	addLaunchFlags(cmd, actor)
	addChartFlags(cmd, &actor.ChartPathOptions, &actor.ChartNameAndDestOptions)
	addValueFlags(cmd.Flags(), &actor.Options)
	return cmd
}

func addLaunchFlags(cmd *cobra.Command, actor *ia.LaunchOpts) {
	cmd.Flags().BoolVar(&actor.DryRun, "dry", false, "simulate an experiment launch")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
}

func init() {
	rootCmd.AddCommand(newLaunchCmd(kd, os.Stdout))
}
