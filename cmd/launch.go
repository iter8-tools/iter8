package cmd

import (
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
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
		Use:   "launch",
		Short: "Launch an experiment",
		Long:  launchDesc,
		Run: func(_ *cobra.Command, _ []string) {
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

// addLaunchFlags adds flags to the launch command
func addLaunchFlags(cmd *cobra.Command, actor *ia.LaunchOpts) {
	cmd.Flags().BoolVar(&actor.DryRun, "dry", false, "simulate an experiment launch")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
}

func init() {
	rootCmd.AddCommand(newLaunchCmd(kd))
}
