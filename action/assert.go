package action

import (
	"fmt"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/action"
)

const (
	// Completed states that the experiment is complete
	Completed = "completed"
	// NoFailure states that none of the tasks in the experiment have failed
	NoFailure = "nofailure"
	// SLOs states that all app versions participating in the experiment satisfy SLOs
	SLOs = "slos"
)

type Assert struct {
	Timeout    time.Duration
	Conditions []string
	// applicable only for kubernetes experiments
	ExperimentResource
}

func NewAssert(cfg *action.Configuration) *Assert {
	return &Assert{}
}

func (assert *Assert) RunKubernetes() (bool, error) {
	if eio, err := assert.ExperimentResource.newKubeOps(); err == nil {
		return assert.Run(eio)
	} else {
		return false, err
	}
}

func (assert *Assert) RunLocal() (bool, error) {
	return assert.Run(&fileOps{})
}

func (assert *Assert) Run(eio base.ExpOps) (bool, error) {
	e, err := build(true, eio)
	if err != nil {
		return false, err
	}

	allGood, err := assert.verify(e)
	if err != nil {
		return false, err
	}
	if !allGood {
		log.Logger.Error("assert conditions failed")
		return false, nil
	}
	return true, nil
}

func (assert *Assert) verify(exp *base.Experiment) (bool, error) {
	// timeSpent tracks how much time has been spent so far in assert attempts
	var timeSpent, _ = time.ParseDuration("0s")

	// sleepTime specifies how long to sleep in between retries of asserts
	var sleepTime, _ = time.ParseDuration("3s")

	// check assert conditions
	allGood := true
	for {
		for _, cond := range assert.Conditions {
			if strings.ToLower(cond) == Completed {
				c := exp.Completed()
				allGood = allGood && c
				if c {
					log.Logger.Info("experiment completed")
				} else {
					log.Logger.Info("experiment did not complete")
				}
			} else if strings.ToLower(cond) == NoFailure {
				nf := exp.NoFailure()
				allGood = allGood && nf
				if nf {
					log.Logger.Info("experiment has no failure")
				} else {
					log.Logger.Info("experiment failed")
				}
			} else if strings.ToLower(cond) == SLOs {
				slos := exp.SLOs()
				allGood = allGood && slos
				if slos {
					log.Logger.Info("SLOs are satisfied")
				} else {
					log.Logger.Info("SLOs are not satisfied")
				}
			} else {
				log.Logger.Error("unsupported assert condition detected; ", cond)
				return false, fmt.Errorf("unsupported assert condition detected; %v", cond)
			}
		}
		if allGood {
			log.Logger.Info("all conditions were satisfied")
			return true, nil
		} else {
			if timeSpent >= assert.Timeout {
				log.Logger.Info("not all conditions were satisfied")
				return false, nil
			} else {
				log.Logger.Infof("sleeping %v ................................", sleepTime)
				time.Sleep(sleepTime)
				timeSpent += sleepTime
			}
		}
	}

}
