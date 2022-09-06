package cmd

import (
	"github.com/iter8-tools/iter8/autox"
	"github.com/spf13/cobra"
)

// autoxDesc is the description of autox cmd
const autoxDesc = `
Run the Iter8 autoX controller.
	iter8 autox
`

// newAutoXCmd creates the autox command
func newAutoXCmd() *cobra.Command {
	// actor := ia.NewAutoXOpts(kd)

	cmd := &cobra.Command{
		Use:   "autox",
		Short: "Start the Iter8 autoX controller",
		Long:  autoxDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			autox.Start()

			return nil
		},
		SilenceUsage: true,
		Hidden:       true,
	}
	return cmd
}

// initialize with assert
func init() {
	rootCmd.AddCommand(newAutoXCmd())
}
