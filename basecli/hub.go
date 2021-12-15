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
package basecli

import (
	"path"

	"github.com/hashicorp/go-getter"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Iter8Hub = "github.com/iter8-tools/iter8.git//hub"
)

var hubFolder string

var hubUsage = `
	Download an experiment folder from the Iter8 hub. 
	This is useful for fetching experiments to inspect, modify, run, or repackage. 
	By default, this command looks for the specified experiment folder in the public Iter8 hub. 
	It is also possible to use custom hubs by setting the ITER8HUB environment variable.

	Environment variables:

	| Name               | Description |
	|--------------------| ------------|
	| $ITER8HUB          | Iter8 hub location. Default value: github.com/iter8-tools/iter8.git//hub |

	The Iter8 hub location follows the following syntax:

	HOST/OWNER/REPO[?ref=branch]//path-to-experiment-folder-relative-to-root-of-the-repo

	For example: github.com/iter8-tools/iter8.git?ref=master//hub
`

// hubCmd represents the hub command
var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "Download an experiment folder from Iter8 hub",
	Long:  hubUsage,
	Example: `
	# download load-test experiment folder from the public Iter8 hub
	iter8 hub -e load-test

	# custom Iter8 hubs are simply github repos that host Iter8 experiment folders
	# Suppose you forked github.com/iter8-tools/iter8, 
	# created a branch called 'ml', and pushed a new experiment folder 
	# called 'tensorflow' under the path 'mkdocs/docs/hub'. 
	# It can now be downloaded as follows.

	export ITER8HUB=github.com/iter8-tools/iter8.git?ref=ml//hub
	iter8 hub -e tensorflow
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// initialize the location of iter8hub
		viper.BindEnv("ITER8HUB")
		viper.SetDefault("ITER8HUB", Iter8Hub)
		ifurl := path.Join(viper.GetString("ITER8HUB"), hubFolder)
		log.Logger.Info("downloading ", ifurl)
		if err := getter.Get(hubFolder, ifurl); err != nil {
			log.Logger.WithStackTrace(err.Error()).Fatalf("unable to get: %v", ifurl)
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(hubCmd)
	hubCmd.Flags().StringVarP(&hubFolder, "experiment", "e", "", "valid experiment folder located under hub")
	hubCmd.MarkFlagRequired("experiment")
}
