package action

import (
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
)

// AutoXOpts are the options used for launching experiments
type AutoXOpts struct {
	// RunOpts provides options relating to experiment resources
	RunOpts
}

// NewAutoXOpts initializes and returns autoX opts
func NewAutoXOpts(kd *driver.KubeDriver) *AutoXOpts {
	return &AutoXOpts{
		RunOpts: *NewRunOpts(kd),
	}
}

// LocalRun launches a local experiment
func (lOpts *AutoXOpts) LocalRun() error {
	log.Logger.Debug("AutoX called")
	return nil
}
