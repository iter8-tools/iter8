package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const kLaunchDesc = `
This command launches an Iter8 experiment in Kubernetes. 

		$ iter8 k launch -c load-test-http --set url=https://httpbin.org/get

To locally render the Kubernetes experiment manifest without running the experiment, use the dry option.

$	iter8 k launch -c load-test-http \
	--set url=https://httpbin.org/get \
	--dry

By default, experiments belong to the 'default' experiment group. To explicitly set the group, use the --group or -g option.

		$	iter8 k launch -c load-test-http \
		--set url=https://httpbin.org/get \
		-g hello

By default, the current directory is used to download and unpack the experiment chart. Control this using the destDir option.

	$	iter8 k launch -c load-test-http \
		--set url=https://httpbin.org/get \
		--destDir /tmp
	
By default, the launch command downloads charts from the official Iter8 chart repo. It is also possible to use third party (helm) repos to host Iter8 experiment charts.

		$	iter8 k launch -c load-test-http \
			--repoURL https://great.expectations.pip \
			--set url=https://httpbin.org/get
`

func newKLaunchCmd() *cobra.Command {
	actor := ia.NewLaunchOpts()

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
	kCmd.AddCommand(newKLaunchCmd())
}
