package cmd

import (
	"os"

	"github.com/iter8-tools/iter8/base/log"
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
	if err := kCmd.PersistentFlags().MarkHidden("debug"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}
	if err := kCmd.PersistentFlags().MarkHidden("registry-config"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}
	if err := kCmd.PersistentFlags().MarkHidden("repository-config"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}
	if err := kCmd.PersistentFlags().MarkHidden("repository-cache"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}

	// add k assert
	kCmd.AddCommand(newKAssertCmd(kd))

	// add k delete
	kCmd.AddCommand(newKDeleteCmd(kd, os.Stdout))

	// add k launch
	kCmd.AddCommand(newKLaunchCmd(kd, os.Stdout))

	// add k log
	kCmd.AddCommand(newKLogCmd(kd))

	// add k report
	kCmd.AddCommand(newKReportCmd(kd))

	// add k run
	if cmd, err := newKRunCmd(kd, os.Stdout); err != nil {
		os.Exit(1)
	} else {
		kCmd.AddCommand(cmd)
	}

}
