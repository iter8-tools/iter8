package cmd

import (
	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"

	"github.com/spf13/cobra"
)

// autoxDesc is the description of autox cmd
const autoxDesc = `
Run the Iter8 AutoX service.
	iter8 autox
`

// newAutoXCmd creates the autox command
func newAutoXCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewAutoXOpts(kd)

	cmd := &cobra.Command{
		Use:   "autox",
		Short: "Start the Iter8 AutoX service",
		Long:  autoxDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun()
		},
		SilenceUsage: true,
		Hidden:       true,
	}
	return cmd
}

// initialize with assert
func init() {
	rootCmd.AddCommand(newAutoXCmd(kd))
}
