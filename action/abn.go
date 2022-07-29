package action

import (
	"github.com/iter8-tools/iter8/abn"
	"github.com/iter8-tools/iter8/driver"
)

// AbmOpts are the options used for launching experiments
type AbnOpts struct {
	// RunOpts provides options relating to experiment resources
	RunOpts
}

// NewAbnOpts initializes and returns abn opts
func NewAbnOpts(kd *driver.KubeDriver) *AbnOpts {
	return &AbnOpts{
		RunOpts: *NewRunOpts(kd),
	}
}

// LocalRun launches a local experiment
func (lOpts *AbnOpts) LocalRun() error {
	abn.Start(lOpts.KubeDriver)
	return nil
}
