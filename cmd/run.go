package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

var runDesc = `
This command runs an experiment specified in the experiment.yaml file and outputs the result of the experiment in the results.yaml file.

		$	iter8 run

This command is primarily intended for development and testing of Iter8 experiment charts and tasks. For production usage, the iter8 launch command is recommended.
`

func newRunCmd() *cobra.Command {
	actor := ia.NewRunOpts()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run an experiment",
		Long:  runDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := actor.LocalRun()
			if err != nil {
				log.Logger.Error(err)
				return err
			}
			return nil
		},
	}
	addRunFlags(cmd, actor)
	return cmd
}

func addRunFlags(cmd *cobra.Command, actor *ia.RunOpts) {
	cmd.Flags().StringVar(&actor.RunDir, "runDir", ".", "directory where experiment is run; contains experiment.yaml and result.yaml")
}

func init() {
	rootCmd.AddCommand(newRunCmd())
}
