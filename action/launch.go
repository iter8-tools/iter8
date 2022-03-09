package action

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
)

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
	return &LaunchOpts{}
}

func (launch *LaunchOpts) LocalRun() error {
	// download chart from Iter8 hub
	if err := launch.HubOpts.Run(); err != nil {
		return err
	}
	// gen experiment spec
	launch.SourceDir = launch.DestDir
	if err := launch.GenOpts.LocalRun(); err != nil {
		return err
	}
	// all done if this is a dry run
	if launch.DryRun {
		return nil
	}
	// run experiment locally
	launch.RunDir = launch.DestDir
	return launch.RunOpts.LocalRun()
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
