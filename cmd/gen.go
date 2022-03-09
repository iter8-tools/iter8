package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const genDesc = `
This command generates an experiment.yaml file by combining an experiment chart with values.

    $ iter8 gen --sourceDir /path/to/load-test-http --set url=https://httpbin.org/get

This command is primarily intended for development and testing of Iter8 experiment charts. For production usage, the launch subcommand is recommended.
`

func newGenCmd() *cobra.Command {
	actor := ia.NewGenOpts()

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate experiment.yaml file by combining an experiment chart with values",
		Long:  genDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := actor.LocalRun()
			if err != nil {
				log.Logger.Error(err)
			}
			return err
		},
	}
	addSourceDirFlag(cmd, &actor.SourceDir, true)
	addValueFlags(cmd.Flags(), &actor.Options)
	return cmd
}

func addSourceDirFlag(cmd *cobra.Command, sourceDirPtr *string, required bool) {
	cmd.Flags().StringVar(sourceDirPtr, "sourceDir", "", "directory where experiment chart resides")
	if required {
		cmd.MarkFlagRequired("sourceDir")
	}
}

func init() {
	rootCmd.AddCommand(newGenCmd())
}
