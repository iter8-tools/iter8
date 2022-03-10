package cmd

import (
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const kAssertDesc = `
This command asserts if the result of a Kubernetes experiment satisfies a given set of conditions.

If the conditions are satisfied, the command exits with code 0. Else, the command exits with code 1.

Assertions are especially useful within CI/CD/GitOps pipelines.

Supported conditions are 'completed', 'nofailure', 'slos', which indicate that the experiment has completed, none of the tasks have failed, and SLOs are satisfied by (all versions of) the app.

    $ iter8 k assert -c completed -c nofailure -c slos

You can optionally specify the group to which the Kubernetes experiment belongs. You can also optionally specify a timeout, which is the maximum amount of time Iter8 should wait in order for the given conditions to be satisfied:

		$ iter8 k assert -c completed,nofailures,slos -t 5s -g hello
`

func newKAssertCmd() *cobra.Command {
	actor := ia.NewAssertOpts()

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
	addExperimentGroupFlag(cmd, &actor.Group, true)
	actor.EnvSettings = settings
	cmd.MarkFlagRequired("namespace")
	addAssertFlags(cmd, actor)
	return cmd
}

func init() {
	kCmd.AddCommand(newKAssertCmd())
}
