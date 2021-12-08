package main

// import (
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/pflag"
// )

// var RootCmd *cobra.Command

// func main() {
// 	flags := pflag.NewFlagSet("iter8", pflag.ExitOnError)
// 	pflag.CommandLine = flags

// 	RootCmd.Execute()

// 	// // extend base CLI root command
// 	// RootCmd := basecli.RootCmd
// 	// // // Add k command
// 	// // rootCmd.AddCommand(k8scli.NewKCommand())
// 	// RootCmd.AddCommand(k8scli.KCmd)
// 	// if err := RootCmd.Execute(); err != nil {
// 	// 	os.Exit(1)
// 	// }
// }

import "github.com/iter8-tools/iter8/k8s/cmd"

func main() {
	cmd.Execute()
}
