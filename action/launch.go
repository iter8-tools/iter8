package action

import (
	"path"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
)

// LaunchOpts are the options used for launching experiments
type LaunchOpts struct {
	// DryRun enables simulating a launch
	DryRun bool
	// HubOpts determines which chart is used and how
	HubOpts
	// GenOpts enables values
	GenOpts
	// RunOpts determines how the experiment is run
	RunOpts
}

// NewHubOpts initializes and returns launch opts
func NewLaunchOpts(kd *driver.KubeDriver) *LaunchOpts {
	hOpts := NewHubOpts()
	rOpts := NewRunOpts(kd)
	return &LaunchOpts{
		HubOpts: *hOpts,
		RunOpts: *rOpts,
	}
}

// LocalRun launches a local experiment
func (lOpts *LaunchOpts) LocalRun() error {
	log.Logger.Debug("launch local run started...")
	// download chart from Iter8 hub
	if err := lOpts.HubOpts.LocalRun(); err != nil {
		return err
	}
	log.Logger.Debug("hub complete")
	// gen experiment spec
	lOpts.GenOpts.SourceDir = path.Join(lOpts.HubOpts.DestDir, lOpts.ChartName)
	log.Logger.Trace("experiment dir: ", lOpts.HubOpts.DestDir)
	log.Logger.Trace("experiment chart dir: ", lOpts.GenOpts.SourceDir)
	if err := lOpts.GenOpts.LocalRun(); err != nil {
		return err
	}
	log.Logger.Debug("gen complete")
	// all done if this is a dry run
	if lOpts.DryRun {
		return nil
	}
	log.Logger.Info("starting local experiment")
	// run experiment locally
	return lOpts.RunOpts.LocalRun()
}

// KubeRun launches a Kubernetes experiment
func (lOpts *LaunchOpts) KubeRun() error {
	// initialize kube driver
	if err := lOpts.KubeDriver.Init(); err != nil {
		return err
	}

	if lOpts.Revision > 0 { // last release found; setup upgrade
		return lOpts.KubeDriver.Upgrade(lOpts.Version, lOpts.ChartName, lOpts.Options, lOpts.Group, lOpts.DryRun, &lOpts.ChartPathOptions)
	} else { // no release found; setup install
		return lOpts.KubeDriver.Install(lOpts.Version, lOpts.ChartName, lOpts.Options, lOpts.Group, lOpts.DryRun, &lOpts.ChartPathOptions)
	}
}
