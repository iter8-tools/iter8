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
	failure      = "failure"
	winnerFound  = "winnerfound"
	winnerPrefix = "winner"
	validPrefix  = "valid"
)

var conds []string
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
		for _, cond := range conds {
			if strings.ToLower(cond) == completed {
				allGood = allGood && exp.completed()
			} else if strings.ToLower(cond) == noFailure {
				allGood = allGood && exp.noFailure()
			} else if strings.ToLower(cond) == failure {
				allGood = allGood && (!exp.noFailure())
			} else if strings.ToLower(cond) == winnerFound {
				allGood = allGood && exp.winnerFound()
			} else if strings.HasPrefix(cond, winnerPrefix) {
				version, err := exp.extractVersion(cond)
				if err != nil {
					os.Exit(1)
				}
				allGood = allGood && exp.isWinner(version)
			} else if strings.HasPrefix(cond, validPrefix) {
				version, err := exp.extractVersion(cond)
				if err != nil {
					os.Exit(1)
				}
				allGood = allGood && exp.isValid(version)
			} else {
				log.Logger.Error("unsupported assert condition detected; ", cond)
				os.Exit(1)
			}
		}
		if allGood {
			log.Logger.Info("all conditions were satisfied")
		} else {
			log.Logger.Info("not all conditions were satisfied")
			os.Exit(1)
		}
	},
}

// completed returns true if the experiment is complete
func (exp *experiment) completed() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.NumCompletedTasks == len(exp.Tasks) {
				log.Logger.Info("experiment completed")
				return true
			}
		}
	}
	log.Logger.Info("experiment did not complete")
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
		return "", errors.New("number of app versions is yet to be initialized")
	}
	for i := 0; i < *exp.Result.NumAppVersions; i++ {
		if tokens[1] == fmt.Sprintf("v%v", i) {
			return tokens[1], nil
		}
	}
	log.Logger.Error("num app versions: ", *exp.Result.NumAppVersions, " invalid version: ", tokens[1])
	return "", errors.New(fmt.Sprint("num app versions: ", *exp.Result.NumAppVersions, " invalid version: ", tokens[1]))
}

func init() {
	rootCmd.AddCommand(assertCmd)
	assertCmd.Flags().StringSliceVarP(&conds, "condition", "c", nil, "completed | noFailure | failure | winnerFound | winner=<version name> | valid=<version name>")
	assertCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "timeout duration (e.g., 5s)")
}

// noFailure returns true if experiment has a results stanza and has not failed
func (exp *experiment) noFailure() bool {
	if exp != nil {
		if exp.Result != nil {
			if !exp.Result.Failure {
				log.Logger.Info("experiment has no failure")
				return true
			}
		}
	}
	log.Logger.Info("experiment failed")
	return false
}

// winnerFound returns true if experiment has a found a winner
func (exp *experiment) winnerFound() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				if exp.Result.Analysis.Winner != nil {
					log.Logger.Info("experiment found a winner")
					return true
				}
			}
		}
	}
	log.Logger.Info("experiment did not find a winner")
	return false
}

// isWinner returns true if version is the winner
func (exp *experiment) isWinner(ver string) bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				if *exp.Result.Analysis.Winner == ver {
					log.Logger.Info("winner is ", ver)
					return true
				}
			}
		}
	}
	log.Logger.Info("winner is not ", ver)
	return false
}

// isValid returns true if version is valid
func (exp *experiment) isValid(ver string) bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				for _, version := range exp.Result.Analysis.Valid {
					if version == ver {
						log.Logger.Info(ver, " is a valid version")
						return true
					}
				}
			}
		}
	}
	log.Logger.Info(ver, " is not a valid version")
	return false
}
