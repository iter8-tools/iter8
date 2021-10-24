package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TestingPatternType identifies the type of experiment type
type TestingPatternType string

const (
	// TestingPatternSLOValidation indicates an experiment tests for SLO validation
	TestingPatternSLOValidation TestingPatternType = "SLOValidation"

	// TestingPatternAB indicates an experiment is a A/B experiment
	TestingPatternAB TestingPatternType = "A/B"

	// TestingPatternABN indicates an experiment is a A/B/n experiment
	TestingPatternABN TestingPatternType = "A/B/N"

	// TestingPatternHybridAB indicates an experiment is a Hybrid-A/B experiment
	TestingPatternHybridAB TestingPatternType = "Hybrid-A/B"

	// TestingPatternHybridABN indicates an experiment is a Hybrid-A/B/n experiment
	TestingPatternHybridABN TestingPatternType = "Hybrid-A/B/N"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "describe an experiment",
	Long:  `Describe an experiment, including the stage of the experiment, how versions are performing with respect to the experiment criteria (reward, objectives, indicators), and information about the winning version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("describe called")
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
