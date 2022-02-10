package cmd

import (
	"github.com/iter8-tools/iter8/base/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logLevel = "info"

var globalUsage = `Safely rollout new versions of apps and ML models. Maximize business value.
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter8",
	Short: "Kubernetes Release Optimizer",
	Long:  globalUsage,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ll, err := logrus.ParseLevel(logLevel)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
		log.Logger.Level = ll
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// disable completion command for now
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log level", "l", "info", "Log level: trace, debug, info (default), warning, error, fatal, panic")
}
