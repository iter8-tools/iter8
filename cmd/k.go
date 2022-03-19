package cmd

import (
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// kCmd is the root command that enables Kubernetes experiments
var kCmd = &cobra.Command{
	Use:   "k",
	Short: "Work with Kubernetes experiments",
	Long:  "Work with Kubernetes experiments",
}

// addExperimentGroupFlag adds the experiment group flag and marks it as required or optional
func addExperimentGroupFlag(cmd *cobra.Command, groupP *string, required bool) {
	cmd.Flags().StringVarP(groupP, "group", "g", driver.DefaultExperimentGroup, "name of the experiment group")
	if required {
		cmd.MarkFlagRequired("group")
	}
}

// addExperimentRevisionFlag adds the experiment revision flag and marks it as required or optional
func addExperimentRevisionFlag(cmd *cobra.Command, revisionP *int, required bool) {
	cmd.Flags().IntVar(revisionP, "revision", 0, "experiment revision")
	if required {
		cmd.MarkFlagRequired("revision")
	}
}

func init() {
	settings.AddFlags(kCmd.PersistentFlags())
	// hiding these Helm flags for now
	kCmd.PersistentFlags().MarkHidden("debug")
	kCmd.PersistentFlags().MarkHidden("registry-config")
	kCmd.PersistentFlags().MarkHidden("repository-config")
	kCmd.PersistentFlags().MarkHidden("repository-cache")
	rootCmd.AddCommand(kCmd)
}
