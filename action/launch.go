package action

import (
	"errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/values"
)

type Launch struct {
	DryRun bool
	Hub
	Gen
	Run
	// applicable only for kubernetes experiments
	Group string
}

func NewLaunch(cfg *action.Configuration) *Launch {
	return &Launch{}
}

func (launch *Launch) RunLocal() error {
	// download chart from Iter8 hub
	if err := launch.download(); err != nil {
		return err
	}
	// gen experiment spec
	if err := launch.gen(); err != nil {
		return err
	}
	if launch.DryRun { // all done
		return nil
	}
	// run experiment locally
	return launch.RunLocal()
}

/*******************
********************

Kubernetes stuff below

********************
********************/

func (launch *Launch) RunKubernetes(values *values.Options) error {
	return errors.New("not implemented")
}
