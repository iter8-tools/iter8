package action

import (
	"errors"
	"path"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
)

const DefaultIter8RepoURL = "https://iter8-tools.github.io/hub"

type LaunchOpts struct {
	DryRun bool
	HubOpts
	GenOpts
	// applicable only for local experiments
	RunOpts
	// applicable only for Kubernetes experiments
	driver.KubeDriver
}

func NewLaunchOpts() *LaunchOpts {
	return &LaunchOpts{
		RunOpts: *NewRunOpts(),
	}
}

func (lOpts *LaunchOpts) LocalRun() error {
	// download chart from Iter8 hub
	if err := lOpts.HubOpts.Run(); err != nil {
		return err
	}
	// gen experiment spec
	lOpts.GenOpts.SourceDir = path.Join(lOpts.HubOpts.DestDir, lOpts.ChartName)
	log.Logger.Trace("experiment dir: ", lOpts.HubOpts.DestDir)
	log.Logger.Trace("experiment chart dir: ", lOpts.GenOpts.SourceDir)
	if err := lOpts.GenOpts.LocalRun(); err != nil {
		return err
	}
	// all done if this is a dry run
	if lOpts.DryRun {
		return nil
	}
	// run experiment locally
	lOpts.RunDir = lOpts.DestDir
	return lOpts.RunOpts.LocalRun()
}

func (lOpts *LaunchOpts) KubeRun() error {
	// initialize kube driver
	if err := lOpts.KubeDriver.Init(); err != nil {
		e := errors.New("unable to initialize KubeDriver")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	if lOpts.Revision > 0 { // last release found; setup upgrade
		return lOpts.KubeDriver.Upgrade(lOpts.Version, lOpts.ChartName, lOpts.Options, lOpts.Group, lOpts.DryRun, &lOpts.ChartPathOptions)
	} else { // no release found; setup install
		return lOpts.KubeDriver.Install(lOpts.Version, lOpts.ChartName, lOpts.Options, lOpts.Group, lOpts.DryRun, &lOpts.ChartPathOptions)
	}
}
