package action

import (
	"errors"

	"github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/values"
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

func NewLaunch(cfg *action.Configuration) *LaunchOpts {
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

/*******************
********************

Kubernetes stuff below

********************
********************/

func (lOpts *LaunchOpts) KubeRun(values *values.Options) error {
	return errors.New("not implemented")
}
