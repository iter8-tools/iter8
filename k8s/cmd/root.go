package cmd

import (
	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
)

var RootCmd = basecli.RootCmd

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	// extend base gen command with the k8s command
	basecli.GenCmd.AddCommand(NewGetK8sCmd())
}
