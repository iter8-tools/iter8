package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

const kLaunchDesc = `
Launch an Iter8 experiment in Kubernetes. 

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

func newKLaunchCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "launch an experiment in Kubernetes",
		Long:  kLaunchDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if err := actor.KubeRun(); err != nil {
				log.Logger.Error(err)
			}
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings
	addLaunchFlags(cmd, actor)
	addChartFlags(cmd, &actor.ChartPathOptions, &actor.ChartNameAndDestOptions)
	addValueFlags(cmd.Flags(), &actor.Options)
	return cmd
}

func init() {
	kCmd.AddCommand(newKLaunchCmd(kd, os.Stdout))
}
