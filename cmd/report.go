package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "report results from experiment",
	Long:  `Report the results of an experiment, including the stage of the experiment, how versions are performing with respect to the experiment criteria (reward, objectives, indicators), and information about the winning version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("report called")
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
