/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
	describeCmd.Flags().StringVarP(&resultFile, "results", "r", "results.yaml", "experiment results yaml file")
}
