package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "report",
	Short: "report results of an experiment",
	Long:  `Report the results of an experiment, including the stage of the experiment, how versions are performing with respect to the experiment criteria (reward, objectives, indicators), and information about the winning version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("report called")
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
