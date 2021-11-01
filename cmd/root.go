package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter8",
	Short: "Metrics driven experiments",
	Example: `
	# run the experiment defined in the local file named experiment.yaml
	iter8 run

	# assert that the experiment completed successfully and found a winner
	iter8 assert -c completed -c successful -c winnerFound
	
	# report experiment results
	iter8 report`,
	SilenceUsage: true,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// disable completion command for now
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
