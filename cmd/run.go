package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

var runDesc = `
Run an experiment specified in experiment.yaml output result to result.yaml.

	$ iter8 run

This command is intended for development and testing of experiment charts and tasks. For production usage, the iter8 launch command is recommended.
`

// newRunCmd creates the run command
func newRunCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewRunOpts(kd)

	cmd := &cobra.Command{
		Use:          "run",
		Short:        "Run an experiment",
		Long:         runDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun()
		},
	}
	addRunDirFlag(cmd, &actor.RunDir)
	return cmd
}

// addRunDirFlag adds run dir flag to the command
func addRunDirFlag(cmd *cobra.Command, runDirPtr *string) {
	cmd.Flags().StringVar(runDirPtr, "runDir", ".", "directory where experiment is run; contains experiment.yaml and result.yaml")
}

func init() {
	rootCmd.AddCommand(newRunCmd(kd, os.Stdout))
}
