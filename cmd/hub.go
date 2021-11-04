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
	gg "github.com/hashicorp/go-getter"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var hubFolder string

var validHubFolders = map[string]bool{
	"load-test": true,
}

// hubCmd represents the hub command
var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "download an experiment folder from Iter8hub",
	Example: `
	# download the load-test experiment folder
	iter8 hub -e load-test
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := validHubFolders[hubFolder]; !ok {
			log.Logger.Errorf("invalid hub folder specified; %s", hubFolder)
			return
		}

		// initialize the location of iter8hub
		viper.BindEnv("ITER8HUB")
		viper.SetDefault("ITER8HUB", "github.com/sriumcp/iter8/mkdocs/docs/iter8hub")
		ifurl := viper.GetString("ITER8HUB") + "/" + hubFolder
		u, err := urlhelper.Parse(ifurl)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Fatalf("unable to parse iter8hub url: %s", ifurl)
			return
		}
		g := new(gg.GitGetter)
		if err := g.Get(hubFolder, u); err != nil {
			log.Logger.WithStackTrace(err.Error()).Fatalf("unable to get: %v", u)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(hubCmd)
	hubCmd.Flags().StringVarP(&hubFolder, "experiment", "e", "", "valid iter8hub folder; must be one of { load-test }")
	hubCmd.MarkFlagRequired("experiment")
}
