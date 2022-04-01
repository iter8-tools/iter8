package action

import (
	"errors"

	"github.com/hashicorp/go-getter"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

var DefaultGitFolder = "github.com/iter8-tools/iter8.git?ref=" + base.Version + "//" + chartsFolderName

// HubOpts are the options used for downloading Iter8 experiment charts
type HubOpts struct {
	// GitFolder is the full path to the GitHub Iter8 experiment charts folder
	GitFolder string
	// ChartsParentDir is the directory where `charts` is to be downloaded
	ChartsParentDir string
}

// NewHubOpts initializes and returns hub opts
func NewHubOpts() *HubOpts {

	return &HubOpts{
		GitFolder:       DefaultGitFolder,
		ChartsParentDir: ".",
	}
}

// LocalRun downloads an experiment chart to DestDir
func (hub *HubOpts) LocalRun() error {
	log.Logger.Infof("downloading %v into %v", hub.GitFolder, hub.ChartsParentDir)
	if err := getter.Get(hub.ChartsParentDir, hub.GitFolder); err != nil {
		e := errors.New("unable to download chart")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	return nil
}
