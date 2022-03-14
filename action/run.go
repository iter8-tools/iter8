package action

import (
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
)

type RunOpts struct {
	// applicable only for local experiments
	RunDir string
	// applicable only for kubernetes experiments
	*driver.KubeDriver
}

func NewRunOpts(kd *driver.KubeDriver) *RunOpts {
	return &RunOpts{
		RunDir:     ".",
		KubeDriver: kd,
	}
}

func (rOpts *RunOpts) LocalRun() error {
	return base.RunExperiment(&driver.FileDriver{
		RunDir: rOpts.RunDir,
	})
}

func (rOpts *RunOpts) KubeRun() error {
	return base.RunExperiment(rOpts)
}
