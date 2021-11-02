package cmd

import (
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
			&base.Experiment{},
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
			return exp.Result.NumCompletedTasks == len(exp.Spec.Tasks)
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
	for _, ver := range exp.Spec.Versions {
		if ver == tokens[1] {
			return ver, nil
		}
	}
	log.Logger.Error("no such version; ", tokens[1])
	return "", fmt.Errorf("no such version; %v", tokens[1])
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
			return !exp.Result.Failure
		}
	}
	return false
}

// winnerFound returns true if experiment has a found a winner
func (exp *experiment) winnerFound() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				return exp.Result.Analysis.Winner != nil
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
				return *exp.Result.Analysis.Winner == ver
			}
		}
	}
	return false
}

// isValid returns true if version is valid
func (exp *experiment) isValid(ver string) bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Analysis != nil {
				for _, version := range exp.Result.Analysis.Valid {
					if version == ver {
						return true
					}
				}
			}
		}
	}
	return false
}
