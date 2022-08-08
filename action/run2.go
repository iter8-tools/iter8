package action

import (
	"github.com/iter8-tools/iter8/base"
)

// RunOpts are the options used for running an experiment
type BaseRunOpts struct {
	// Rundir is the directory of the local experiment.yaml file
	RunDir string

	// KubeDriver enables Kubernetes experiment run
	*base.KubeDriver

	// ReuseResult configures Iter8 to reuse the experiment result instead of
	// creating a new one for looping experiments.
	ReuseResult bool
}

// NewRunOpts initializes and returns run opts
func NewBaseRunOpts(kd *base.KubeDriver) *BaseRunOpts {
	return &BaseRunOpts{
		RunDir:     ".",
		KubeDriver: kd,
	}
}

// KubeRun runs a Kubernetes experiment
func (rOpts *BaseRunOpts) KubeRun() error {
	// initialize kube driver
	if err := rOpts.KubeDriver.InitKube(); err != nil {
		return err
	}
	return base.RunExperiment(rOpts.ReuseResult, rOpts.KubeDriver)
}
