package action

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
)

type ChartNameAndDestOptions struct {
	ChartName string
	DestDir   string
}

type HubOpts struct {
	ChartNameAndDestOptions
	action.ChartPathOptions
}

func NewHubOpts() *HubOpts {
	return &HubOpts{}
}

// clean pre-existing chart artifacts in destination dir
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
		log.Logger.Info("removed ", f)
	}
	return nil
}

// Run downloads an experiment chart
func (hub *HubOpts) Run() error {
	// removing any pre-existing files and dirs matching the glob
	if err := hub.cleanChartArtifacts(); err != nil {
		return err
	}
	log.Logger.Info("cleaned up any existing chart artifacts")

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

	log.Logger.Infof("pulling %v from %v into %v", hub.ChartName, pull.RepoURL, pull.DestDir)
	_, err = pull.Run(hub.ChartName)
	if err != nil {
		e := fmt.Errorf("unable to get %v", hub.ChartName)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	return nil
}
