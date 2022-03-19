package action

import (
	"github.com/iter8-tools/iter8/driver"
)

// LogOpts enables fetching logs from Kubernetes
type LogOpts struct {
	// KubeDriver enables interaction with Kubernetes cluster
	*driver.KubeDriver
}

// NewHubOpts initializes and returns log opts
func NewLogOpts(kd *driver.KubeDriver) *LogOpts {
	return &LogOpts{
		KubeDriver: kd,
	}
}

// KubeRun fetches logs from a Kubernetes experiment
func (lOpts *LogOpts) KubeRun() (string, error) {
	if err := lOpts.KubeDriver.Init(); err != nil {
		return "", err
	}
	return lOpts.GetExperimentLogs()
}
