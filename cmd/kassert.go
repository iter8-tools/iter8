package cmd

import (
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

const kAssertDesc = `
Assert if the result of a Kubernetes experiment satisfies a given set of conditions. If all conditions are satisfied, the command exits with code 0. Else, the command exits with code 1. 

Assertions are especially useful for automation inside CI/CD/GitOps pipelines.

Supported conditions are 'completed', 'nofailure', 'slos', which indicate that the experiment has completed, none of the tasks have failed, and the SLOs are satisfied.

	$ iter8 k assert -c completed -c nofailure -c slos
	# same as iter8 k assert -c completed,nofailure,slos

You can optionally specify a timeout, which is the maximum amount of time to wait for the conditions to be satisfied:

	$ iter8 k assert -c completed,nofailures,slos -t 5s
`

func newKAssertCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewAssertOpts(kd)

	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Assert if Kubernetes experiment result satisfies conditions",
		Long:  kAssertDesc,
		Run: func(_ *cobra.Command, _ []string) {
			allGood, err := actor.KubeRun()
			if err != nil {
				log.Logger.Error(err)
			}
			if !allGood {
				log.Logger.Error("assert conditions failed")
				os.Exit(1)
			}
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings
	addAssertFlags(cmd, actor)
	return cmd
}

func init() {
	kCmd.AddCommand(newKAssertCmd(kd))
}
