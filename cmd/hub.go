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
	"os"
	"path"
	"path/filepath"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
)

const (
	defaultIter8RepoURL = "https://iter8-tools.github.io/hub"
)

var (
	chartName              string
	chartVersionConstraint string
	repoURL                string
	destDir                string
)

var hubUsage = `
Download an experiment chart from an Iter8 experiment chart repo.
This is useful for fetching experiments to inspect, modify, launch, or repackage. 
By default, this command looks for the specified experiment chart in the default Iter8 experiment repo. You can use third party repos by supplying the repo URL flag.

The default Iter8 experiment chart repo has the following URL:
https://iter8-tools.github.io/hub
`

// hubCmd represents the hub command
var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "Download an experiment chart from an Iter8 experiment chart repo",
	Long:  hubUsage,
	Example: `
# download the load-test-http experiment chart from 
# the default Iter8 experiment chart repo
iter8 hub -c load-test-http

# download the great-expectations experiment chart from 
# the third party experiment chart repo whose URL is 
# https://great.expectations.pip
iter8 hub -c great-expectations -r https://great.expectations.pip
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// set up helm pull object
		cfg := &action.Configuration{
			Capabilities: chartutil.DefaultCapabilities,
		}
		pull := action.NewPullWithOpts(action.WithConfig(cfg))
		pull.Settings = cli.New()
		pull.Untar = true
		pull.RepoURL = repoURL
		pull.Version = chartVersionConstraint
		if pull.Version == "" {
			pull.Version = string(base.MajorMinor) + ".x"
		}

		var err error
		pull.DestDir = destDir
		pull.UntarDir = destDir

		log.Logger.Infof("pulling %v from %v into %v", chartName, pull.RepoURL, pull.DestDir)
		_, err = pull.Run(chartName)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Errorf("unable to get %v", chartName)
			os.Exit(1)
		}
		return nil
	},
}

// clean pre-existing chart artifacts in destination dir
func cleanChartArtifacts(d string, c string) error {
	// removing any pre-existing files and dirs matching the glob
	files, err := filepath.Glob(path.Join(destDir, chartName+"*"))
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			log.Logger.Error(err)
			return err
		}
		log.Logger.Info("removed ", f)
	}
	return nil
}

func init() {
	hubCmd.Flags().StringVarP(&chartName, "chartName", "c", "", "name of the experiment chart")
	hubCmd.MarkFlagRequired("chartName")
	hubCmd.Flags().StringVarP(&chartVersionConstraint, "chartVersionConstraint", "v", "", "version constraint for chart (example 0.1.0)")
	hubCmd.Flags().StringVarP(&repoURL, "repoURL", "r", defaultIter8RepoURL, "URL of experiment chart repo")
	hubCmd.Flags().StringVarP(&destDir, "destDir", "d", ".", "destination folder where experiment chart is downloaded and unpacked; by default, Iter8 will create and use a temporary folder as destination")

	rootCmd.AddCommand(hubCmd)

}
