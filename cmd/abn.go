package cmd

import (
	"os"
	"os/signal"
	"syscall"

	abn "github.com/iter8-tools/iter8/abn/core"
	"github.com/spf13/cobra"
)

// abnDesc is the description of abn cmd
const abnDesc = `
Run the Iter8 A/B(/n) service.

	iter8 abn
`

// port number on which gRPC service should listen
var port int

// newAbnCmd creates the abn command
func newAbnCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "abn",
		Short: "Start the Iter8 A/B(/n) service",
		Long:  abnDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			stopCh := make(chan struct{})
			defer close(stopCh)
			if err := abn.Start(port, stopCh); err != nil {
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
	addPortFlag(cmd, &port)
	return cmd
}

// addTimeoutFlag adds timeout flag to command
func addPortFlag(cmd *cobra.Command, portPtr *int) {
	cmd.Flags().IntVar(portPtr, "port", 50051, "service port")
}
