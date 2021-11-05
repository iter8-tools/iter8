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

	# assert that the experiment completed without failure and found a winner
	iter8 assert -c completed -c noFailure -c winnerFound
	
	# report experiment results using the built-in text template
	iter8 gen

	# report experiment results using a custom go template specified in iter8.tpl file
	iter8 gen -o custom`,
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
