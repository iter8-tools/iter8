package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// controllersDesc is the description of controllers cmd
const controllersDesc = `
Start Iter8 controllers.

	iter8 controllers
`

// port number on which A/B/n gRPC service listens
var port int

// newControllersCmd creates the Iter8 controllers
// when invoking this function for real, set stopCh to nil
// this will block the controller from exiting until an os.Interrupt;
// when invoking this function in a unit test, set stopCh to ctx.Done()
// this will exit the controller when cancel() is called by the parent function;
func newControllersCmd(stopCh <-chan struct{}, client k8sclient.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "controllers",
		Short:        "Start Iter8 controllers",
		Long:         controllersDesc,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			// createSigCh indicates if we should create sigCh (channel) that fires on interrupt
			createSigCh := false

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// if stopCh is nil, create sigCh to exit this func on interrupt,
			// and use ctx.Done() to clean up controllers when exiting;
			// otherwise, simply use stopCh for both
			if stopCh == nil {
				stopCh = ctx.Done()
				createSigCh = true
			}

			if client == nil {
				var err error
				client, err = k8sclient.New(settings)
				if err != nil {
					log.Logger.Error("could not obtain Kube client ... ")
					return err
				}
			}

			if err := controllers.Start(stopCh, client); err != nil {
				log.Logger.Error("controllers did not start ... ")
				return err
			}
			log.Logger.Debug("started controllers ... ")

			// launch gRPC server to respond to frontend requests
			go controllers.LaunchGRPCServer(port, []grpc.ServerOption{}, stopCh)

			// if createSigCh, then block until there is an os.Interrupt
			if createSigCh {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
				<-sigCh
				log.Logger.Warn("SIGTERM ... ")
			}

			return nil
		},
	}
	return cmd
}

// addTimeoutFlag adds timeout flag to command
func addPortFlag(cmd *cobra.Command, portPtr *int) {
	cmd.Flags().IntVar(portPtr, "port", 50051, "service port")
}
