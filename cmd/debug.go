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

var priority uint8

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "debug an experiment",
	Long:  `Print logs for an experiment. Logs will be at the given priority level or higher.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("debug called")
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().Uint8VarP(&priority, "priority", "p", 1, "1, 2, or 3 corresponding to high, medium and low; for example, setting priority to 2 would print logs of priority 1 or 2")
	debugCmd.Flags().StringVarP(&logFile, "log", "l", "experiment.log", "experiment log file")
}
