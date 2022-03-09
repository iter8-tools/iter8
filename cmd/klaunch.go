package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/values"
)

const kLaunchDesc = `
This command launches an Iter8 experiment in Kubernetes. 

		$ iter8 k launch -c load-test-http --set url=https://httpbin.org/get

To locally render the Kubernetes experiment manifest without running the experiment, use the dry option.

$	iter8 k launch -c load-test-http \
	--set url=https://httpbin.org/get \
	--dry

By default, the current directory is used to download and unpack the experiment chart. Control this using the destDir option.

	$	iter8 k launch -c load-test-http \
		--set url=https://httpbin.org/get \
		--destDir /tmp
	
By default, the launch command downloads charts from the official Iter8 chart repo. It is also possible to use third party (helm) repos to host Iter8 experiment charts.

		$	iter8 k launch -c load-test-http \
			--repoURL https://great.expectations.pip \
			--set url=https://httpbin.org/get
`

func newKLaunchCmd(cfg *action.Configuration) *cobra.Command {
	actor := ia.NewLaunch(cfg)
	valueOpts := &values.Options{}

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "launch an experiment in Kubernetes",
		Long:  kLaunchDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := actor.KubeRun(valueOpts)
			if err != nil {
				log.Logger.Error(err)
				return err
			}
			return nil
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	addLaunchFlags(cmd, actor)
	return cmd
}
