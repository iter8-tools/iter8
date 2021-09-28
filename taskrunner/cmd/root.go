package cmd

import (
	"fmt"
	"os"

	"github.com/iter8-tools/etc3/taskrunner/core"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// config
var cfgFile string

// log
var log *logrus.Logger

// package variables used for holding flag values
var action string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "handler",
	Short: "perform start up and completion tasks in iter8 experiments",
	Long:  `iter8 launches jobs at the start and completition of an experiment, and executes the handler program within the job's containers.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".handler.yaml", "config file (default is .handler.yaml)")
	log = core.GetLogger()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".handler" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".handler")
	}

	// Read in any config file
	viper.ReadInConfig()

	viper.AutomaticEnv() // read in environment variables that match

	ll, err := logrus.ParseLevel(viper.GetString("log_level"))
	if err == nil {
		core.SetLogLevel(ll)
	}
}
