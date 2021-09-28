package cmd

import (
	"fmt"
	"os"

	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var expName string
var expNamespace string
var latest bool
var exp *expr.Experiment
var priority uint8

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter8ctl",
	Short: "Iter8 command line utility",
	Long:  `iter8ctl promotes understanding of an Iter8 experiment. It can be used to describe the stage of the experiment, how versions are performing, and assert various conditions relating to the experiment. This program is a K8s client and requires a valid K8s cluster with Iter8 installed.`,
	Args:  cobra.MaximumNArgs(1),
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iter8ctl.yaml)")

	rootCmd.PersistentFlags().StringVarP(&expNamespace, "namespace", "n", "", "namespace of the experiment; never ignored -- namespace from current context is used if not specified")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".iter8ctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".iter8ctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
