package cmd

import (
	"errors"

	"github.com/iter8-tools/etc3/iter8ctl/describe"
	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/spf13/cobra"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe [experiment-name]",
	Short: "Describe an Iter8 experiment",
	Long:  `Summarize an experiment, including the stage of the experiment, how versions are performing with respect to the experiment criteria (reward, SLOs, metrics), and information about the winning version. When experiment-name is omitted, the experiment with the latest creation timestamp in the cluster is described.`,
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
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		describe.Builder().WithExperiment(exp).PrintAnalysis()
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// describeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// describeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
