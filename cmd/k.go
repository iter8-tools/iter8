package cmd

import (
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli"
)

var (
	settings = cli.New()
)

var kCmd = &cobra.Command{
	Use:   "k",
	Short: "Work with Kubernetes experiments",
	Long:  "Work with Kubernetes experiments",
}

func addExperimentGroupFlag(cmd *cobra.Command, groupP *string, required bool) {
	cmd.Flags().StringVarP(groupP, "group", "g", driver.DefaultExperimentGroup, "name of the experiment group")
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
	settings.AddFlags(kCmd.PersistentFlags())
	rootCmd.AddCommand(kCmd)
}
