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

	// ReuseResult configures Iter8 to reuse the experiment result instead of
	// creating a new one for looping experiments.
	ReuseResult bool
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
	}, rOpts.ReuseResult)
}

// KubeRun runs a Kubernetes experiment
func (rOpts *RunOpts) KubeRun() error {
	// initialize kube driver
	if err := rOpts.KubeDriver.InitKube(); err != nil {
		return err
	}
	return base.RunExperiment(rOpts.KubeDriver, rOpts.ReuseResult)
}
