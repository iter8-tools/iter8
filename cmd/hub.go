package cmd

import (
	ia "github.com/iter8-tools/iter8/action"

	"github.com/spf13/cobra"
)

// hubDesc is the description of the hub command
const hubDesc = `
Download Iter8 experiment charts from a Git repo or other sources.

	$ iter8 hub

This command is intended for development and testing of experiment charts. For production usage, the iter8 launch command is recommended.
`

// newHubCmd creates the hub command
func newHubCmd() *cobra.Command {
	actor := ia.NewHubOpts()

	cmd := &cobra.Command{
		Use:          "hub",
		Short:        "Download experiment charts",
		Long:         hubDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun()
		},
	}
	addRemoteFolderURLFlag(cmd, &actor.RemoteFolderURL)
	return cmd
}

// add the remoteFolderURL flag to the command
func addRemoteFolderURLFlag(cmd *cobra.Command, remoteFolderURLPtr *string) {
	cmd.Flags().StringVar(remoteFolderURLPtr, "remoteFolderURL", ia.DefaultRemoteFolderURL(), "URL of the remote folder containing Iter8 experiment charts. Accepts any URL supported by https://github.com/hashicorp/go-getter")
}

// initialize with the hub command
func init() {
	rootCmd.AddCommand(newHubCmd())
}
