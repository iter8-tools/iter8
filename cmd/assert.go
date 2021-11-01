package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/core"
	"github.com/iter8-tools/iter8/core/log"
	"github.com/spf13/cobra"
)

const (
	// Completed implies experiment is complete
	Completed = "completed"
	NoFailure = "nofailure"
	Failure   = "failure"

	// WinnerFound implies experiment has found a winner
	WinnerFound = "winnerfound"

	// WinnerPrefix
	WinnerPrefix = "winner"

	// ValidPrefix
	ValidPrefix = "valid"
)

var conds []string
var timeout time.Duration

// assertCmd represents the assert command
var assertCmd = &cobra.Command{
	Use:   "assert",
	Short: "assert if the experiment satisfies the specified conditions",
	Run: func(cmd *cobra.Command, args []string) {
		// build experiment
		exp := &Experiment{
			&core.Experiment{},
		}
		log.Logger.Trace("build started")
		exp, err := Build(true)
		log.Logger.Trace("build finished")
		if err != nil {
			log.Logger.Error("experiment build failed")
			os.Exit(1)
		}

		// check assert conditions
		allGood := true
		for _, cond := range conds {
			if strings.ToLower(cond) == Completed {
				allGood = allGood && exp.Completed()
			} else if strings.ToLower(cond) == NoFailure {
				allGood = allGood && exp.noFailure()
			} else if strings.ToLower(cond) == Failure {
				allGood = allGood && (!exp.noFailure())
			} else if strings.ToLower(cond) == WinnerFound {
				allGood = allGood && exp.winnerFound()
			} else if strings.HasPrefix(cond, WinnerPrefix) {
				version, err := exp.extractVersion(cond)
				if err != nil {
					os.Exit(1)
				}
				allGood = allGood && exp.isWinner(version)
			} else if strings.HasPrefix(cond, ValidPrefix) {
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

// Completed returns true if the experiment is complete
func (exp *Experiment) Completed() bool {
	if exp != nil {
		if exp.Result != nil {
			return exp.Result.NumCompletedTasks == len(exp.Spec.Tasks)
		}
	}
	return false
}

// extract version from string
func (exp *Experiment) extractVersion(cond string) (string, error) {
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
func (exp *Experiment) noFailure() bool {
	if exp != nil {
		if exp.Result != nil {
			return !exp.Result.Failure
		}
	}
	return false
}

// winnerFound returns true if experiment has a found a winner
func (exp *Experiment) winnerFound() bool {
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
func (exp *Experiment) isWinner(ver string) bool {
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
func (exp *Experiment) isValid(ver string) bool {
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
