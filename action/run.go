package action

import (
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
)

// RunOpts are the options used for running an experiment
type RunOpts struct {
	// Rundir is the directory of the local experiment.yaml file
	RunDir string
	// KubeDriver enables Kubernetes experiment run
	*driver.KubeDriver
}

// NewRunOpts initializes and returns run opts
func NewRunOpts(kd *driver.KubeDriver) *RunOpts {
	return &RunOpts{
		RunDir:     ".",
		KubeDriver: kd,
	}
}

// LocalRun runs a local experiment
func (rOpts *RunOpts) LocalRun() error {
	return base.RunExperiment(&driver.FileDriver{
		RunDir: rOpts.RunDir,
	})
}

// KubeRun runs a Kubernetes experiment
func (rOpts *RunOpts) KubeRun() error {
	// initialize kube driver
	if err := rOpts.KubeDriver.Init(); err != nil {
		return err
	}
	return base.RunExperiment(rOpts)
}
