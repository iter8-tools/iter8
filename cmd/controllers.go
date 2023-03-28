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
)

// controllersDesc is the description of controllers cmd
const controllersDesc = `
Start Iter8 controllers.

	iter8 controllers
`

// newControllersCmd creates the Iter8 controllers
func newControllersCmd(client k8sclient.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "controllers",
		Short:        "Start Iter8 controllers",
		Long:         controllersDesc,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if client == nil {
				var err error
				client, err = k8sclient.New(settings)
				if err != nil {
					log.Logger.Error("could not obtain Kube client ... ")
					return err
				}
			}

			if err := controllers.Start(ctx.Done(), client); err != nil {
				log.Logger.Error("controllers did not start ... ")
				return err
			}
			log.Logger.Debug("started controllers ... ")

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
			<-sigCh

			log.Logger.Warn("SIGTERM ... ")

			return nil
		},
	}
	return cmd
}
