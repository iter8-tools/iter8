package cmd

import (
	"fmt"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const kLogDesc = `
This command fetches the logs for a Kubernetes experiment.

		$	iter8 k log

or 

		$	iter8 k log --group hello
`

func newKLogCmd() *cobra.Command {
	actor := ia.NewLogOpts()

	cmd := &cobra.Command{
		Use:   "log",
		Short: "get logs for a Kubernetes experiment",
		Long:  kLogDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			lg, err := actor.KubeRun()
			if err != nil {
				log.Logger.Error(err)
				return err
			}
			fmt.Println("experiment logs...")
			fmt.Println(lg)
			return nil
		},
	}
	addExperimentGroupFlag(cmd, &actor.Group, false)
	actor.EnvSettings = settings
	return cmd
}

func init() {
	kCmd.AddCommand(newKLogCmd())
}
