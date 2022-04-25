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
	// Folder is the full path to the Iter8 experiment charts folder
	Folder string
	// ChartsDir is the full path to the `charts` dir
	ChartsDir string
}

func DefaultFolder() string {
	// parse version
	v := strings.Split(base.Version, "-")
	ref := v[0]
	return defaultIter8Repo + "?ref=" + ref + "//" + chartsFolderName
}

// NewHubOpts initializes and returns hub opts
func NewHubOpts() *HubOpts {

	return &HubOpts{
		Folder:    DefaultFolder(),
		ChartsDir: chartsFolderName,
	}
}

// LocalRun downloads an experiment chart to DestDir
func (hub *HubOpts) LocalRun() error {
	log.Logger.Infof("downloading %v into %v", hub.Folder, hub.ChartsDir)
	if err := getter.Get(hub.ChartsDir, hub.Folder); err != nil {
		e := errors.New("unable to download charts")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	return nil
}
