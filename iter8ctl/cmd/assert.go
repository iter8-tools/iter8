package cmd

import (
	"errors"
	"fmt"
	"os"

	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/spf13/cobra"
)

var conditions []string
var conds []expr.ConditionType

// assertCmd represents the assert command
var assertCmd = &cobra.Command{
	Use:   "assert [experiment-name]",
	Short: "Assert conditions for an Iter8 experiment",
	Long:  `One or more conditions can be asserted using this command for an Iter8 experiment. This command is especially useful in CI/CD/Gitops pipelines prior to version promotion or rollback. When experiment-name is omitted, the experiment with the latest creation timestamp in the cluster is used for assertions.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("more than one positional argument supplied")
		}
		latest = (len(args) == 0)
		if !latest {
			expName = args[0]
		}
		// at this stage, either latest must be true or expName must be non-empty
		if !latest && expName == "" {
			panic("either latest must be true or expName must be non-empty")
		}
		// get experiment from cluster
		var err error
		if exp, err = expr.GetExperiment(latest, expName, expNamespace); err != nil {
			return err
		}
		// parse conditions
		if len(conditions) == 0 {
			return errors.New("one or more conditions must be specified with assert")
		}
		for _, cond := range conditions {
			switch cond {
			case string(expr.Completed):
				conds = append(conds, expr.Completed)
			case string(expr.WinnerFound):
				conds = append(conds, expr.WinnerFound)
			default:
				return errors.New("Invalid condition: " + cond)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := exp.Assert(conds); err == nil {
			fmt.Println("All conditions satisfied.")
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(assertCmd)
	assertCmd.Flags().StringSliceVarP(&conditions, "condition", "c", nil, "completed | winnerFound")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// assertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// assertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
