/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
	core.Logger.Info("LOG_LEVEL ", ll)
	core.SetLogLevel(ll)
}
