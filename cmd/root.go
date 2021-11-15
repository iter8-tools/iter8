package cmd

import (
	"github.com/spf13/cobra"
)

var globalUsage = `The Kubernetes metrics-driven experimentation platform

Environment variables:

| Name               | Description |
|--------------------| ------------|
| $LOG_LEVEL         | Iter8 log level. Values are: Trace, Debug, Info (default), Warning, Error, Fatal and Panic. |
`

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "iter8",
	Short:   "Metrics driven experiments",
	Long:    globalUsage,
	Version: "v0.8",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	// disable completion command for now
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.InitDefaultVersionFlag()
}
