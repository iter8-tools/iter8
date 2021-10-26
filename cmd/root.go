package cmd

import (
	"github.com/iter8-tools/iter8/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter8",
	Short: "Metrics driven experiments",
	Example: `
	# view the current namespace in your KUBECONFIG
	ns

	# view all of the namespaces in use by contexts in your KUBECONFIG
	ns --list
	
	# switch your current-context to one that contains the desired namespace
	ns foo`,
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
	// initialize log level
	viper.BindEnv("LOG_LEVEL")
	viper.SetDefault("LOG_LEVEL ", "info")
	ll, _ := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	core.Logger.Debug("LOG_LEVEL ", ll)
	core.SetLogLevel(ll)
}
