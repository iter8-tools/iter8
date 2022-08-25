package cmd

import (
	"github.com/iter8-tools/iter8/abn"
	"github.com/iter8-tools/iter8/driver"

	"github.com/spf13/cobra"
)

// abnDesc is the description of abn cmd
const abnDesc = `
Run the Iter8 A/B(/n) service.

	iter8 abn
`

// newAbnCmd creates the abn command
func newAbnCmd(kd *driver.KubeDriver) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "abn",
		Short: "Start the Iter8 A/B(/n) service",
		Long:  abnDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			abn.Start(kd)
			return nil
		},
		SilenceUsage: true,
		Hidden:       true,
	}
	return cmd
}

// initialize with assert
func init() {
	rootCmd.AddCommand(newAbnCmd(kd))
}
