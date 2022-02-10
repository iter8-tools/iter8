package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logLevel = uint32(logrus.InfoLevel)

var globalUsage = `Safely rollout new versions of apps and ML models. Maximize business value.
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter8",
	Short: "Kubernetes Release Optimizer",
	Long:  globalUsage,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// disable completion command for now
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().Uint32VarP(&logLevel, "log level", "l", uint32(logrus.InfoLevel), "Log level")

}
