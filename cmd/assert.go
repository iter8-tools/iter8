package cmd

import (
	"fmt"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const assertDesc = `
This command asserts if the result of an experiment satisfies a given set of conditions.

If the conditions are satisfied, the command exits with code 0. Else, the command exits with code 1. 

Assertions are especially useful within CI/CD/GitOps pipelines.

Supported conditions are 'completed', 'nofailure', 'slos', which indicate that the experiment has completed, none of the tasks have failed, and SLOs are satisfied by (all versions of) the app.

    $ iter8 assert -c completed -c nofailure -c slos

or

    $ iter8 assert -c completed,nofailure,slos

You can optionally specify a timeout, which is the maximum amount of time Iter8 should wait in order for the given conditions to be satisfied:

		$ iter8 assert -c completed,nofailures,slos -t 5s
`

func newAssertCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewAssertOpts(kd)

	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Assert if experiment result satisfies conditions",
		Long:  assertDesc,
		Run: func(_ *cobra.Command, _ []string) {
			allGood, err := actor.LocalRun()
			if err != nil {
				log.Logger.Error(err)
			}
			if !allGood {
				log.Logger.Error("assert conditions failed")
				os.Exit(1)
			}
		},
	}
	addAssertFlags(cmd, actor)
	addRunFlags(cmd, &actor.RunOpts)
	return cmd
}

func addAssertFlags(cmd *cobra.Command, actor *ia.AssertOpts) {
	cmd.Flags().StringSliceVarP(&actor.Conditions, "condition", "c", nil, fmt.Sprintf("%v | %v | %v; can specify multiple or separate conditions with commas;", ia.Completed, ia.NoFailure, ia.SLOs))
	cmd.MarkFlagRequired("condition")
	cmd.Flags().DurationVar(&actor.Timeout, "timeout", 0, "timeout duration (e.g., 5s)")
}

func init() {
	rootCmd.AddCommand(newAssertCmd(kd))
}
