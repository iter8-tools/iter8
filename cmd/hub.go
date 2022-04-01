package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

	"github.com/spf13/cobra"
)

const hubDesc = `
Download an experiment chart to a local directory.

	$ iter8 hub -c load-test-http

This command is intended for development and testing of experiment charts. For production usage, the iter8 launch command is recommended.
`

// newHubCmd creates the hub command
func newHubCmd() *cobra.Command {
	actor := ia.NewHubOpts()

	cmd := &cobra.Command{
		Use:          "hub",
		Short:        "Download experiment chart",
		Long:         hubDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun()
		},
	}
	addGitFolderFlag(cmd, &actor.GitFolder)
	return cmd
}

// addGitFolderFlag
func addGitFolderFlag(cmd *cobra.Command, gitFolderPtr *string) {
	cmd.Flags().StringVar(gitFolderPtr, "gitFolder", ia.DefaultGitFolder, "Git folder containing iter8 charts")
}

func init() {
	rootCmd.AddCommand(newHubCmd())
}
