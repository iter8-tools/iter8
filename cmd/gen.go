package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const genDesc = `
Generate an experiment.yaml file by combining an experiment chart with values.

    $ iter8 gen --sourceDir /path/to/load-test-http --set url=https://httpbin.org/get

This command is intended for development and testing of experiment charts. For production usage, the launch subcommand is recommended.
`

func newGenCmd() *cobra.Command {
	actor := ia.NewGenOpts()

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate experiment.yaml file by combining an experiment chart with values",
		Long:  genDesc,
		Run: func(_ *cobra.Command, _ []string) {
			if err := actor.LocalRun(); err != nil {
				log.Logger.Error(err)
			}
		},
	}
	addSourceDirFlag(cmd, &actor.SourceDir, true)
	addValueFlags(cmd.Flags(), &actor.Options)
	return cmd
}

func addSourceDirFlag(cmd *cobra.Command, sourceDirPtr *string, required bool) {
	cmd.Flags().StringVar(sourceDirPtr, "sourceDir", "", "path to experiment chart directory")
	if required {
		cmd.MarkFlagRequired("sourceDir")
	}
}

func init() {
	rootCmd.AddCommand(newGenCmd())
}
