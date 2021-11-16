/*
Copyright © 2021 Iter8 authors

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
	"github.com/hashicorp/go-getter"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var hubFolder string

var hubUsage = `
Download an experiment folder from Iter8 Hub. This is useful for fetching experiments to inspect, modify, run or repackage. By default, this command looks for the specified experiment folder in the public Iter8 hub. It is also possible to use private hubs by setting the ITER8HUB environment variable.

Environment variables:

| Name               | Description |
|--------------------| ------------|
| $ITER8HUB          | Iter8 hub location. Default value: github.com/iter8-tools/iter8.git//mkdocs/docs/hub/ |
`

// hubCmd represents the hub command
var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "download an experiment folder from Iter8 Hub",
	Long:  hubUsage,
	Example: `
	# public hub
	iter8 hub -e load-test

	# private hub
	# Suppose you forked github.com/iter8-tools/iter8, 
	# created a branch called 'ml', and pushed a new experiment folder 
	# called 'tensorflow' under the path 'mkdocs/docs/hub'. 
	# It can now be downloaded as follows.

	export ITER8HUB=github.com/iter8-tools/iter8.git?ref=ml//mkdocs/docs/hub/
	iter8 hub -e tensorflow
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// initialize the location of iter8hub
		viper.BindEnv("ITER8HUB")
		viper.SetDefault("ITER8HUB", "github.com/iter8-tools/iter8.git//mkdocs/docs/hub/")
		ifurl := viper.GetString("ITER8HUB") + hubFolder
		if err := getter.Get(hubFolder, ifurl); err != nil {
			log.Logger.WithStackTrace(err.Error()).Fatalf("unable to get: %v", ifurl)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(hubCmd)
	hubCmd.Flags().StringVarP(&hubFolder, "experiment", "e", "", "valid experiment folder located under hub")
	hubCmd.MarkFlagRequired("experiment")
}
