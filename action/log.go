package action

import (
	"github.com/iter8-tools/iter8/driver"
)

type LogOpts struct {
	driver.KubeDriver
}

func NewLogOpts() *LogOpts {
	return &LogOpts{}
}

func (lOpts *LogOpts) KubeRun() (string, error) {
	if err := lOpts.KubeDriver.Init(); err != nil {
		return "", err
	}
	return lOpts.GetExperimentLogs()
}
