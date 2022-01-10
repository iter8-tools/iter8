package basecli

import (
	"github.com/spf13/cobra"
)

var globalUsage = `Safely rollout new versions of apps and ML models. Maximize business value with each release.

Environment variables:

| Name               | Description |
|--------------------| ------------|
| $LOG_LEVEL         | Iter8 log level. Values are: Trace, Debug, Info (default), Warning, Error, Fatal and Panic. |
`

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "iter8",
	Short: "Kubernetes release engineering",
	Long:  globalUsage,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	// disable completion command for now
	RootCmd.CompletionOptions.DisableDefaultCmd = true
}
