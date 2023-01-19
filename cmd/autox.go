package cmd

import (
	"os"
	"os/signal"
	"syscall"

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
	cmd := &cobra.Command{
		Use:   "autox",
		Short: "Start the Iter8 autoX controller",
		Long:  autoxDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			stopCh := make(chan struct{})
			defer close(stopCh)

			if err := autox.Start(stopCh, nil); err != nil {
				return err
			}
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
			<-sigCh

			return nil
		},
		SilenceUsage: true,
		Hidden:       true,
	}
	return cmd
}
