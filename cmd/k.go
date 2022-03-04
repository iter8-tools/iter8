package cmd

import (
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli"
)

var settings = cli.New()

var kCmd = &cobra.Command{
	Use:   "k",
	Short: "Work with Kubernetes experiments",
	Long:  "Work with Kubernetes experiments",
}

func init() {
	rootCmd.AddCommand(kCmd)
	flags := kCmd.PersistentFlags()
	settings.AddFlags(flags)
}
