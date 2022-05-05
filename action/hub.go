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
	// RemoteFolderURL is the URL of the remote Iter8 experiment charts folder
	// Remote URLs can be any go-getter URLs like GitHub or GitLab URLs
	// https://github.com/hashicorp/go-getter
	RemoteFolderURL string
	// ChartsDir is the full path to the `charts` dir
	ChartsDir string
}

func DefaultRemoteFolderURL() string {
	// parse version
	v := strings.Split(base.Version, "-")
	ref := v[0]
	return defaultIter8Repo + "?ref=" + ref + "//" + charts
}

// NewHubOpts initializes and returns hub opts
func NewHubOpts() *HubOpts {

	return &HubOpts{
		RemoteFolderURL: DefaultRemoteFolderURL(),
		ChartsDir:       charts,
	}
}

// LocalRun downloads an experiment chart to DestDir
func (hub *HubOpts) LocalRun() error {
	log.Logger.Infof("downloading %v into %v", hub.RemoteFolderURL, hub.ChartsDir)
	if err := getter.Get(hub.ChartsDir, hub.RemoteFolderURL); err != nil {
		e := errors.New("unable to download charts")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	return nil
}
