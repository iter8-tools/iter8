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

func addExperimentGroupFlag(cmd *cobra.Command, group *string, required bool) {
	cmd.Flags().StringVarP(group, "group", "g", defaultExperimentGroup, "name of the experiment group")
	if required {
		cmd.MarkFlagRequired("group")
	}
}

func init() {
	rootCmd.AddCommand(kCmd)
	flags := kCmd.PersistentFlags()
	settings.AddFlags(flags)
}
