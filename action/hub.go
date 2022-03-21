package action

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
)

// ChartNameAndDestOptions specifies which chart needs to be downloaded and where
type ChartNameAndDestOptions struct {
	// ChartName is the name of the chart
	ChartName string
	// DestDir is the directory where it is to be downloaded
	DestDir string
}

// HubOpts are the options used for downloading an experiment chart
type HubOpts struct {
	// ChartNameAndDestOptions contains ChartName and DestDir options
	ChartNameAndDestOptions
	// ChartPathOptions contains RepoURL and (chart) Version (constraint) options
	action.ChartPathOptions
}

// NewHubOpts initializes and returns hub opts
func NewHubOpts() *HubOpts {
	return &HubOpts{
		ChartNameAndDestOptions: ChartNameAndDestOptions{
			DestDir: ".",
		},
		ChartPathOptions: action.ChartPathOptions{
			RepoURL: driver.DefaultIter8RepoURL,
		},
	}
}

// cleanChartArtifacts cleans any pre-existing chart artifacts in DestDir
func (hub *HubOpts) cleanChartArtifacts() error {
	// removing any pre-existing files and dirs matching the glob
	files, err := filepath.Glob(path.Join(hub.DestDir, hub.ChartName+"*"))
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			log.Logger.Error(err)
			return err
		}
		log.Logger.Debug("removed ", f)
	}
	return nil
}

// LocalRun downloads an experiment chart to DestDir
func (hub *HubOpts) LocalRun() error {
	// removing any pre-existing files and dirs matching the glob
	if err := hub.cleanChartArtifacts(); err != nil {
		return err
	}
	log.Logger.Debug("cleaned up any existing chart artifacts")

	// set up helm pull object
	cfg := &action.Configuration{
		Capabilities: chartutil.DefaultCapabilities,
	}
	pull := action.NewPullWithOpts(action.WithConfig(cfg))
	pull.Settings = cli.New()
	pull.Untar = true
	pull.RepoURL = hub.RepoURL
	pull.Version = hub.Version
	if pull.Version == "" {
		pull.Version = string(base.MajorMinor) + ".x"
	}

	var err error
	pull.DestDir = hub.DestDir
	pull.UntarDir = hub.DestDir

	log.Logger.Debugf("pulling from %v", pull.RepoURL)
	log.Logger.Infof("pulling %v", hub.ChartName)
	log.Logger.Debugf("pulling into %v", pull.DestDir)
	_, err = pull.Run(hub.ChartName)
	if err != nil {
		e := fmt.Errorf("unable to pull %v", hub.ChartName)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	return nil
}
