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

// addTestFlag adds the test flag
func addTestFlag(cmd *cobra.Command, testP *string) {
	cmd.Flags().StringVarP(testP, "test", "t", driver.DefaultTestName, "name of the test")
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

	// add k run
	kcmd.AddCommand(newKRunCmd(kd, os.Stdout))
}
