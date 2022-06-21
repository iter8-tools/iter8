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

// addExperimentGroupFlag adds the experiment group flag
func addExperimentGroupFlag(cmd *cobra.Command, groupP *string) {
	cmd.Flags().StringVarP(groupP, "group", "g", driver.DefaultExperimentGroup, "name of the experiment group")
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
