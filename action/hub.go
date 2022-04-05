package action

import (
	"errors"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

const defaultIter8Repo = "github.com/iter8-tools/iter8.git"

// HubOpts are the options used for downloading Iter8 experiment charts
type HubOpts struct {
	// GitFolder is the full path to the GitHub Iter8 experiment charts folder
	GitFolder string
	// ChartsDir is the full path to the `charts` dir
	ChartsDir string
}

func DefaultGitFolder() string {
	// parse version
	v := strings.Split(base.Version, "-")
	ref := v[0]
	return defaultIter8Repo + "?ref=" + ref + "//" + chartsFolderName
}

// NewHubOpts initializes and returns hub opts
func NewHubOpts() *HubOpts {

	return &HubOpts{
		GitFolder: DefaultGitFolder(),
		ChartsDir: chartsFolderName,
	}
}

// LocalRun downloads an experiment chart to DestDir
func (hub *HubOpts) LocalRun() error {
	log.Logger.Infof("downloading %v into %v", hub.GitFolder, hub.ChartsDir)
	if err := getter.Get(hub.ChartsDir, hub.GitFolder); err != nil {
		e := errors.New("unable to download chart")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	return nil
}
