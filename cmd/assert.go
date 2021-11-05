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
	completed        = "completed"
	noFailure        = "nofailure"
	winnerFound      = "winnerfound"
	winnerPrefix     = "winner"
	satisfyingPrefix = "satisfying"
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
	Short: "assert if the experiment satisfies the specified conditions",
	Run: func(cmd *cobra.Command, args []string) {
		// build experiment
		exp := &experiment{
			Experiment: &base.Experiment{},
		}
		log.Logger.Trace("build started")
		exp, err := build(true)
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
				} else if strings.ToLower(cond) == winnerFound {
					wf := exp.winnerFound()
					allGood = allGood && wf
					if wf {
						log.Logger.Info("experiment found a winner")
					} else {
						log.Logger.Info("experiment did not find a winner")
					}
				} else if strings.HasPrefix(cond, winnerPrefix) {
					version, err := exp.extractVersion(cond)
					var iw bool
					if err != nil {
						iw = false
					} else {
						iw = exp.isWinner(version)
					}
					allGood = allGood && iw
					if iw {
						log.Logger.Info("winner is ", version)
					} else {
						log.Logger.Info("winner is not ", version)
					}
				} else if strings.HasPrefix(cond, satisfyingPrefix) {
					version, err := exp.extractVersion(cond)
					if err != nil {
						os.Exit(1)
					}
					iv := exp.isSatisfying(version)
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
func (exp *experiment) completed() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.NumCompletedTasks == len(exp.Tasks) {
				return true
			}
		}
	}
	return false
}

// extract version from string
func (exp *experiment) extractVersion(cond string) (string, error) {
	tokens := strings.Split(cond, "=")
	if len(tokens) != 2 {
		log.Logger.Error("unsupported condition detected; ", cond)
		return "", fmt.Errorf("unsupported condition detected; %v", cond)
	}
	if exp.Result == nil || exp.Result.NumAppVersions == nil {
		log.Logger.Error("number of app versions is yet to be initialized")
		return tokens[1], errors.New("number of app versions is yet to be initialized")
	}
	for i := 0; i < *exp.Result.NumAppVersions; i++ {
		if tokens[1] == fmt.Sprintf("v%v", i) {
			return tokens[1], nil
		}
	}
	log.Logger.Error("valid version names: ", exp.appVersions(), " invalid version name specified: ", tokens[1])
	return tokens[1], errors.New(fmt.Sprint("num app versions: ", *exp.Result.NumAppVersions, " invalid version: ", tokens[1]))
}

// noFailure returns true if experiment has a results stanza and has not failed
func (exp *experiment) noFailure() bool {
	if exp != nil {
		if exp.Result != nil {
			if !exp.Result.Failure {
				return true
			}
		}
	}
	return false
}

// winnerFound returns true if experiment has a found a winner
func (exp *experiment) winnerFound() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				if exp.Result.Analysis.Winner != nil {
					return true
				}
			}
		}
	}
	return false
}

// isWinner returns true if version is the winner
func (exp *experiment) isWinner(ver string) bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				if *exp.Result.Analysis.Winner == ver {
					return true
				}
			}
		}
	}
	return false
}

// isSatisfying returns true if version satisfies objectives
func (exp *experiment) isSatisfying(ver string) bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				for _, version := range exp.Result.Analysis.Satisfying {
					if version == ver {
						return true
					}
				}
			}
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(assertCmd)
	assertCmd.Flags().StringSliceVarP(&conds, "condition", "c", nil, "completed | noFailure | winnerFound | winner=<version name> | satisfying=<version name>")
	assertCmd.MarkFlagRequired("condition")
	assertCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "timeout duration (e.g., 5s)")
}
