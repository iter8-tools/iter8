package cmd

import (
	"os"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
)

// kcmd is the root command that enables Kubernetes experiments
var kcmd = &cobra.Command{
	Use:   "k",
	Short: "Work with Kubernetes experiments",
	Long:  "Work with Kubernetes experiments",
}

// addExperimentGroupFlag adds the experiment group flag
func addExperimentGroupFlag(cmd *cobra.Command, groupP *string) {
	cmd.Flags().StringVarP(groupP, "group", "g", driver.DefaultExperimentGroup, "name of the experiment group")
}

func init() {
	settings.AddFlags(kcmd.PersistentFlags())
	// hiding these Helm flags for now
	if err := kcmd.PersistentFlags().MarkHidden("debug"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}
	if err := kcmd.PersistentFlags().MarkHidden("registry-config"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}
	if err := kcmd.PersistentFlags().MarkHidden("repository-config"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}
	if err := kcmd.PersistentFlags().MarkHidden("repository-cache"); err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}

	// add k assert
	kcmd.AddCommand(newKAssertCmd(kd))

	// add k delete
	kcmd.AddCommand(newKDeleteCmd(kd, os.Stdout))

	// add k launch
	kcmd.AddCommand(newKLaunchCmd(kd, os.Stdout))

	// add k log
	kcmd.AddCommand(newKLogCmd(kd))

	// add k report
	kcmd.AddCommand(newKReportCmd(kd))

	// add k run
	kcmd.AddCommand(newKRunCmd(kd, os.Stdout))

}
