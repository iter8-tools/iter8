package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/iter8-tools/etc3/controllers"
	"github.com/iter8-tools/etc3/iter8ctl/debug"
	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/spf13/cobra"
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug [experiment-name]",
	Short: "Debug an Iter8 experiment",
	Long:  `Print logs for an Iter8 experiment sorted in chronological order and filtered priority. Currently, debug is restricted to logs from Iter8's task runner jobs. In the future, this will include support for logs from controller and analytics as well.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("more than one positional argument supplied")
		}

		// at this stage, either latest must be true or expName must be non-empty
		latest = (len(args) == 0)
		if !latest {
			expName = args[0]
		}
		if !latest && expName == "" {
			panic("either latest must be true or expName must be non-empty")
		}

		// get experiment from cluster
		var err error
		if exp, err = expr.GetExperiment(latest, expName, expNamespace); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ils, err := debug.Debug(exp, controllers.Iter8LogPriority(priority))
		if err == nil {
			fmt.Printf("Debugging experiment %s in namespace %s\n", exp.Name, exp.Namespace)
			for _, il := range ils {
				fmt.Printf("source: %v priority: %v message: %s\n", il.Source, il.Priority, il.Message)
			}
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.PersistentFlags().Uint8VarP(&priority, "priority", "p", 1, "1, 2, or 3 corresponding to high, medium and low; for example, setting priority to 2 would print logs of priority 1 or 2")
}
