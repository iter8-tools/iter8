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
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v42/github"
	"github.com/hashicorp/go-getter"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// latestStableVersion returns the latest stable version of Iter8
func latestStableVersion() (string, error) {
	// find all tags
	client := github.NewClient(nil)
	tags, _, err := client.Repositories.ListTags(context.Background(), "iter8-tools", "iter8", nil)
	// something went wrong or found zero tags
	if err != nil || len(tags) == 0 {
		e := errors.New("unable to determine latest stable version for Iter8 hub")
		var msg string
		if len(tags) == 0 {
			msg = "found 0 tags"
		}
		msg = e.Error()
		log.Logger.WithStackTrace(msg).Error(e)
	}
	// found some tags
	log.Logger.Infof("found %v tags", len(tags))
	// found latest tag with the correct major minor prefix
	if strings.HasPrefix(*tags[0].Name, majorMinor+".") {
		return *tags[0].Name, nil
	}
	// ToDo: Fix the following error
	err = fmt.Errorf("unable to find tags with major minor %v", majorMinor)
	log.Logger.Error(err)
	return "", err
}

// getIter8Hub gets the location of the Iter8Hub
func getIter8Hub() (string, error) {
	iter8HubTpl := "github.com/iter8-tools/iter8.git?ref=%v//hub"
	viper.BindEnv("ITER8HUB")
	iter8HubFromEnv := viper.GetString("ITER8HUB")
	if len(iter8HubFromEnv) > 0 {
		return iter8HubFromEnv, nil
	}
	tag, err := latestStableVersion()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(iter8HubTpl, tag), nil
}

var hubFolder string

var hubUsage = `
Download an experiment chart from the Iter8 hub. 
This is useful for fetching experiments to inspect, modify, run, or repackage. 
By default, this command looks for the specified experiment chart in the public Iter8 hub. 
It is also possible to use custom hubs by setting the ITER8HUB environment variable.

Environment variables:

| Name               | Description |
|--------------------| ------------|
| $ITER8HUB          | Iter8 hub location. Default value: github.com/iter8-tools/iter8.git//hub |

The Iter8 hub location follows the following syntax:

HOST/OWNER/REPO[?ref=branch]//path-to-experiment-folder-relative-to-root-of-the-repo

For example, the public Iter8 hub is located at:
github.com/iter8-tools/iter8.git?ref=master//hub
`

// hubCmd represents the hub command
var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "Download an experiment chart from Iter8 hub",
	Long:  hubUsage,
	Example: `
# download the load-test-http experiment chart from the public Iter8 hub
iter8 hub -e load-test-http

# custom Iter8 hubs are simply github repos that host Iter8 experiment charts
# Suppose you forked github.com/iter8-tools/iter8 under the GitHub account $GHUSER,
# created a branch called 'ml', and pushed a new experiment chart 
# called 'tensorflow' under the path 'my/path/to/hub'. 
# It can now be downloaded as follows.

export ITER8HUB=github.com/$GHUSER/iter8.git?ref=ml//my/path/to/hub
iter8 hub -e tensorflow
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		hubRoot, err := getIter8Hub()
		if err != nil {
			return err
		}
		ifurl := path.Join(hubRoot, hubFolder)
		log.Logger.Info("downloading ", ifurl)
		if err := getter.Get(hubFolder, ifurl); err != nil {
			log.Logger.WithStackTrace(err.Error()).Errorf("unable to get: %v", ifurl)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(hubCmd)
	hubCmd.Flags().StringVarP(&hubFolder, "experiment", "e", "", "valid experiment chart located under hub")
	hubCmd.MarkFlagRequired("experiment")
}
