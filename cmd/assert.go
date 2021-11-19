package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const (
	completed    = "completed"
	noFailure    = "nofailure"
	slos         = "slos"
	slosByPrefix = "slosby"
)

// assert conditions
var conds []string

// how long to sleep in between retries of asserts
var sleepTime, _ = time.ParseDuration("3s")

// how long have we spent so far in assert attempts
var timeSpent, _ = time.ParseDuration("0s")

// timeout for assert conditions to be satisfied
var timeout time.Duration

// assertCmd represents the assert command
var assertCmd = &cobra.Command{
	Use:   "assert",
	Short: "assert if experiment run satisfies the specified conditions",
	Long:  "Assert if experiment run satisfies the specified conditions. This command exits with code 0, if assert conditions are satisfied. Else, it returns with code 1.",
	Example: `
	# download the load-test experiment
	iter8 hub -e load-test

	cd load-test

	# run it
	iter8 run

	# assert that the experiment completed without failures, 
	# and SLOs were satisfied
	iter8 assert -c completed -c nofailure -c slos

	# another way to write the above assertion
	iter8 assert -c completed,nofailure,slos

	# if the experiment involves multiple app versions, 
	# SLOs can be asserted for individual versions
	# for example, the following command asserts that
	# SLOs are satisfied by version numbered 0
	iter8 assert -c completed,nofailures,slosby=0

	# timeouts are useful for an experiment that may be long running
	# and may run in the background
	iter8 assert -c completed,nofailures,slosby=0 -t 5s
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// build experiment
		exp := &Experiment{
			Experiment: &base.Experiment{},
		}
		log.Logger.Trace("build started")
		fio := &FileExpIO{}
		exp, err := Build(true, fio)
		log.Logger.Trace("build finished")
		if err != nil {
			log.Logger.Error("experiment build failed")
			os.Exit(1)
		}

		// check assert conditions
		allGood := true
		for {
			for _, cond := range conds {
				if strings.ToLower(cond) == completed {
					c := exp.completed()
					allGood = allGood && c
					if c {
						log.Logger.Info("experiment completed")
					} else {
						log.Logger.Info("experiment did not complete")
					}
				} else if strings.ToLower(cond) == noFailure {
					nf := exp.noFailure()
					allGood = allGood && nf
					if nf {
						log.Logger.Info("experiment has no failure")
					} else {
						log.Logger.Info("experiment failed")
					}
				} else if strings.ToLower(cond) == slos {
					slos := exp.SLOs()
					allGood = allGood && slos
					if slos {
						log.Logger.Info("SLOs are satisfied")
					} else {
						log.Logger.Info("SLOs are not satisfied")
					}
				} else if strings.HasPrefix(cond, slosByPrefix) {
					version, err := exp.extractVersion(cond)
					if err != nil {
						os.Exit(1)
					}
					iv := exp.SLOsBy(version)
					allGood = allGood && iv
					if iv {
						log.Logger.Info(version, " satisfies objectives")
					} else {
						log.Logger.Info(version, " does not satisfy objectives")
					}
				} else {
					log.Logger.Error("unsupported assert condition detected; ", cond)
					os.Exit(1)
				}
			}
			if allGood {
				log.Logger.Info("all conditions were satisfied")
				os.Exit(0)
			} else {
				if timeSpent > timeout {
					log.Logger.Info("not all conditions were satisfied")
					os.Exit(1)
				} else {
					time.Sleep(sleepTime)
					timeSpent += sleepTime
				}
			}
		}
	},
}

// completed returns true if the experiment is complete
// if the result stanza is missing, this function returns false
func (exp *Experiment) completed() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.NumCompletedTasks == len(exp.Tasks) {
				return true
			}
		}
	}
	return false
}

// noFailure returns true if no task int he experiment has failed
// if the result stanza is missing, this function returns false
func (exp *Experiment) noFailure() bool {
	if exp != nil {
		if exp.Result != nil {
			if !exp.Result.Failure {
				return true
			}
		}
	}
	return false
}

// extract version from string
func (exp *Experiment) extractVersion(cond string) (int, error) {
	tokens := strings.Split(cond, "=")
	if len(tokens) != 2 {
		log.Logger.Error("unsupported condition detected; ", cond)
		return -1, fmt.Errorf("unsupported condition detected; %v", cond)
	}
	if exp.Result == nil || exp.Result.Insights == nil || exp.Result.Insights.NumAppVersions == nil {
		log.Logger.Error("number of app versions is uninitialized")
		return -1, errors.New("number of app versions is uninitialized")
	}
	for i := 0; i < *exp.Result.Insights.NumAppVersions; i++ {
		if tokens[1] == fmt.Sprintf("%v", i) {
			return i, nil
		}
	}
	log.Logger.Error("number of app versions: ", *exp.Result.Insights.NumAppVersions, "; valid app version must be in the range 0 to ", *exp.Result.Insights.NumAppVersions-1)
	return -1, errors.New(fmt.Sprint("number of app versions: ", *exp.Result.Insights.NumAppVersions, "; valid app version must be in the range 0 to ", *exp.Result.Insights.NumAppVersions-1))
}

func init() {
	RootCmd.AddCommand(assertCmd)
	assertCmd.Flags().StringSliceVarP(&conds, "condition", "c", nil, fmt.Sprintf("%v | %v | %v | %v=<version number>", completed, noFailure, slos, slosByPrefix))
	assertCmd.MarkFlagRequired("condition")
	assertCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "timeout duration (e.g., 5s)")
}
