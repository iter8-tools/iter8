/*
 */

package main

import (
	"os"

	basecli "github.com/iter8-tools/iter8/cmd"
	k8scli "github.com/iter8-tools/iter8/k8s/cmd"

	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("iter8", pflag.ExitOnError)
	pflag.CommandLine = flags

	// extend base CLI root command
	root := basecli.RootCmd
	// Add k command
	root.AddCommand(k8scli.NewCmdKCommand())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
