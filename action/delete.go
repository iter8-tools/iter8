package action

import (
	"github.com/iter8-tools/iter8/driver"
)

// DeleteOpts are the options used for deleting experiment groups
type DeleteOpts struct {
	// KubeDriver enables access to Kubernetes cluster
	*driver.KubeDriver
}

// NewHubOpts initializes and returns launch opts
func NewDeleteOpts(kd *driver.KubeDriver) *DeleteOpts {
	return &DeleteOpts{
		KubeDriver: kd,
	}
}

// KubeRun deletes a Kubernetes experiment
func (dOpts *DeleteOpts) KubeRun() error {
	// initialize kube driver
	if err := dOpts.KubeDriver.Init(); err != nil {
		return err
	}

	return dOpts.KubeDriver.Delete()
}
