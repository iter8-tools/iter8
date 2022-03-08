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

func addExperimentGroupFlag(cmd *cobra.Command, groupP *string, required bool) {
	cmd.Flags().StringVarP(groupP, "group", "g", defaultExperimentGroup, "name of the experiment group")
	if required {
		cmd.MarkFlagRequired("group")
	}
}

func addExperimentRevisionFlag(cmd *cobra.Command, revisionP *int, required bool) {
	cmd.Flags().IntVar(revisionP, "revision", 0, "experiment revision")
	if required {
		cmd.MarkFlagRequired("revision")
	}
}

func init() {
	rootCmd.AddCommand(kCmd)
	flags := kCmd.PersistentFlags()
	settings.AddFlags(flags)
}
