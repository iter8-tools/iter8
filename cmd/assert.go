package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ConditionType is a type for conditions that can be asserted
type ConditionType string

const (
	// Completed implies experiment is complete
	Completed ConditionType = "completed"
	// Successful     ConditionType = "successful"
	// Failure        ConditionType = "failure"
	// HandlerFailure ConditionType = "handlerFailure"

	// WinnerFound implies experiment has found a winner
	WinnerFound ConditionType = "winnerFound"
	// CandidateWon   ConditionType = "candidateWon"
	// BaselineWon    ConditionType = "baselineWon"
	// NoWinner       ConditionType = "noWinner"
)

var conds []string

// assertCmd represents the assert command
var assertCmd = &cobra.Command{
	Use:   "assert",
	Short: "assert if the experiment satisfies the specified conditions",
	Long:  `Assert one or more conditions using this command. Assertions can be used in CI/CD/Gitops pipelines as part of automated version promotion or rollback.`,
	Args: func(cmd *cobra.Command, args []string) error {
		conditions := []ConditionType{}
		for _, cond := range conds {
			switch cond {
			case string(Completed):
				conditions = append(conditions, Completed)
			case string(WinnerFound):
				conditions = append(conditions, WinnerFound)
			default:
				return errors.New("Invalid condition: " + cond)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("assert called")
	},
}

func init() {
	rootCmd.AddCommand(assertCmd)
	assertCmd.Flags().StringSliceVarP(&conds, "condition", "c", nil, "completed | winnerFound")
}
