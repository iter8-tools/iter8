package action

import (
	"fmt"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
)

const (
	// Completed states that the experiment is complete
	Completed = "completed"
	// NoFailure states that none of the tasks in the experiment have failed
	NoFailure = "nofailure"
	// SLOs states that all app versions participating in the experiment satisfy SLOs
	SLOs = "slos"
)

// AssertOpts are the options used for asserting experiment results
type AssertOpts struct {
	// Timeout is the duration to wait for conditions to be satisfied
	Timeout time.Duration
	// Conditions are checked by assert
	Conditions []string
	// RunOpts provides options relating to experiment resources
	RunOpts
}

// NewAssertOpts initializes and returns assert opts
func NewAssertOpts(kd *driver.KubeDriver) *AssertOpts {
	return &AssertOpts{
		RunOpts: *NewRunOpts(kd),
	}
}

// LocalRun asserts conditions for a local experiment
func (aOpts *AssertOpts) LocalRun() (bool, error) {
	return aOpts.Run(&driver.FileDriver{
		RunDir: aOpts.RunDir,
	})
}

// KubeRun asserts conditions for a Kubernetes experiment
func (aOpts *AssertOpts) KubeRun() (bool, error) {
	if err := aOpts.KubeDriver.Init(); err != nil {
		return false, err
	}

	return aOpts.Run(aOpts.KubeDriver)
}

// Run builds the experiment and verifies assert conditions
func (aOpts *AssertOpts) Run(eio base.Driver) (bool, error) {
	allGood, err := aOpts.verify(eio)
	if err != nil {
		return false, err
	}
	if !allGood {
		log.Logger.Error("assert conditions failed")
		return false, nil
	}
	return true, nil
}

// verify implements the core logic of assert
func (aOpts *AssertOpts) verify(eio base.Driver) (bool, error) {
	// timeSpent tracks how much time has been spent so far in assert attempts
	var timeSpent, _ = time.ParseDuration("0s")

	// sleepTime specifies how long to sleep in between retries of asserts
	var sleepTime, _ = time.ParseDuration("3s")

	// check assert conditions
	for {
		exp, err := base.BuildExperiment(eio)
		if err != nil {
			return false, err
		}

		allGood := true

		for _, cond := range aOpts.Conditions {
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
		}
		if timeSpent >= aOpts.Timeout {
			log.Logger.Info("not all conditions were satisfied")
			return false, nil
		}
		log.Logger.Infof("sleeping %v ................................", sleepTime)
		time.Sleep(sleepTime)
		timeSpent += sleepTime
	}

}
