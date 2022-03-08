package action

import (
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
)

type RunOpts struct {
	RunDir string
	// applicable only for kubernetes experiments
	driver.KubeDriver
}

func NewRunOpts() *RunOpts {
	return &RunOpts{
		RunDir: ".",
	}
}

func (runner *RunOpts) LocalRun() error {
	return base.RunExperiment(&driver.FileDriver{
		RunDir: runner.RunDir,
	})
}

func (runner *RunOpts) KubeRun() error {
	if err := runner.KubeDriver.Init(); err != nil {
		return err
	}
	return base.RunExperiment(&runner.KubeDriver)
}
