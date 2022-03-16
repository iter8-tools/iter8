package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

var runDesc = `
This command runs an experiment specified in the experiment.yaml file and outputs the result of the experiment in the results.yaml file.

		$	iter8 run

This command is primarily intended for development and testing of Iter8 experiment charts and tasks. For production usage, the iter8 launch command is recommended.
`

func newRunCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewRunOpts(kd)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run an experiment",
		Long:  runDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if err := actor.LocalRun(); err != nil {
				log.Logger.Error(err)
			}
		},
	}
	addRunFlags(cmd, actor)
	return cmd
}

func addRunFlags(cmd *cobra.Command, actor *ia.RunOpts) {
	cmd.Flags().StringVar(&actor.RunDir, "runDir", ".", "directory where experiment is run; contains experiment.yaml and result.yaml")
}

func init() {
	rootCmd.AddCommand(newRunCmd(kd, os.Stdout))
}
