/*
 */

package main

import (
	"os"

	basecli "github.com/iter8-tools/iter8/cmd"
	k8scli "github.com/iter8-tools/iter8/k8s/pkg/cmd"

	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("iter8", pflag.ExitOnError)
	pflag.CommandLine = flags

	// root := cmd.NewCmdIter8Command()
	root := basecli.RootCmd
	root.AddCommand(k8scli.NewCmdK8sCommand())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
