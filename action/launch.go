package action

import (
	"errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/values"
)

type LaunchOpts struct {
	DryRun bool
	HubOpts
	GenOpts
	RunOpts
	// applicable only for kubernetes experiments
	Group string
}

func NewLaunch(cfg *action.Configuration) *LaunchOpts {
	return &LaunchOpts{}
}

func (launch *LaunchOpts) LocalRun() error {
	// download chart from Iter8 hub
	if err := launch.download(); err != nil {
		return err
	}
	// gen experiment spec
	launch.SourceDir = launch.DestDir
	if err := launch.gen(); err != nil {
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

func (launch *LaunchOpts) KubeRun(values *values.Options) error {
	return errors.New("not implemented")
}
