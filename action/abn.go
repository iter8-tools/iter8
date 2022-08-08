package action

import (
	"github.com/iter8-tools/iter8/abn"
	"github.com/iter8-tools/iter8/base"
)

// AbmOpts are the options used for launching experiments
type AbnOpts struct {
	// RunOpts provides options relating to experiment resources
	BaseRunOpts
}

// NewAbnOpts initializes and returns abn opts
func NewAbnOpts(kd *base.KubeDriver) *AbnOpts {
	return &AbnOpts{
		BaseRunOpts: *NewBaseRunOpts(kd),
	}
}

// LocalRun launches a local experiment
func (lOpts *AbnOpts) LocalRun() error {
	abn.Start(lOpts.KubeDriver)
	return nil
}
