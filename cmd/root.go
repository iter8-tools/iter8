package cmd

import (
	"github.com/iter8-tools/iter8/basecli"
)

var RootCmd = basecli.RootCmd

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.Execute()
}
