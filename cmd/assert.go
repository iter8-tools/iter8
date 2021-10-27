package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/iter8-tools/iter8/core"
	task "github.com/iter8-tools/iter8/tasks"
	"github.com/spf13/cobra"
)

// ConditionType is a type for conditions that can be asserted
type ConditionType string

const (
	// Completed implies experiment is complete
	Completed ConditionType = "completed"
	NoFailure ConditionType = "nofailure"
	Failure   ConditionType = "failure"

	// WinnerFound implies experiment has found a winner
	WinnerFound ConditionType = "winnerfound"

	// WinnerPrefix
	WinnerPrefix ConditionType = "winner"

	// ValidPrefix
	ValidPrefix ConditionType = "valid"
)

var conds []string

// assertCmd represents the assert command
var assertCmd = &cobra.Command{
	Use:   "assert",
	Short: "assert if the experiment satisfies the specified conditions",
	Run: func(cmd *cobra.Command, args []string) {
		// build experiment
		exp := &core.Experiment{
			TaskMaker: &task.TaskMaker{},
		}
		core.Logger.Trace("build started")
		err := exp.Build(true)
		core.Logger.Trace("build finished")
		if err != nil {
			core.Logger.Error("experiment build failed")
			os.Exit(1)
		}

		// check assert conditions
		allGood := true
		for _, cond := range conds {
			if strings.ToLower(cond) == string(Completed) {
				allGood = allGood && exp.Completed()
			} else if strings.ToLower(cond) == string(NoFailure) {
				allGood = allGood && exp.NoFailure()
			} else if strings.ToLower(cond) == string(Failure) {
				allGood = allGood && (!exp.NoFailure())
			} else if strings.ToLower(cond) == string(WinnerFound) {
				allGood = allGood && exp.WinnerFound()
			} else if strings.HasPrefix(cond, string(WinnerPrefix)) {
				version, err := extractVersion(exp, cond)
				if err != nil {
					os.Exit(1)
				}
				allGood = allGood && exp.IsWinner(version)
			} else if strings.HasPrefix(cond, string(ValidPrefix)) {
				version, err := extractVersion(exp, cond)
				if err != nil {
					os.Exit(1)
				}
				allGood = allGood && exp.IsValid(version)
			} else {
				core.Logger.Error("unsupported assert condition detected; ", cond)
				os.Exit(1)
			}
		}
		if allGood {
			fmt.Println("all conditions satisfied")
		} else {
			os.Exit(1)
		}
	},
}

func extractVersion(exp *core.Experiment, cond string) (string, error) {
	tokens := strings.Split(cond, "=")
	if len(tokens) != 2 {
		core.Logger.Error("unsupported condition detected; ", cond)
		return "", fmt.Errorf("unsupported condition detected; %v", cond)
	}
	for _, ver := range exp.Spec.Versions {
		if ver == tokens[1] {
			return ver, nil
		}
	}
	core.Logger.Error("no such version; ", tokens[1])
	return "", fmt.Errorf("no such version; %v", tokens[1])
}

func init() {
	rootCmd.AddCommand(assertCmd)
	assertCmd.Flags().StringSliceVarP(&conds, "condition", "c", nil, "completed | noFailure | failure | winnerFound | winner=<version name> | valid=<version name>")
}
