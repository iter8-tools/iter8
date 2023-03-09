package cmd

import (
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
func newControllersCmd(kClient k8sclient.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controllers",
		Short: "Start Iter8 controllers",
		Long:  abnDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			stopCh := make(chan struct{})
			defer close(stopCh)

			if err := controllers.Start(stopCh, kClient); err != nil {
				log.Logger.Error("controllers did not start ... ")
				return err
			}
			log.Logger.Trace("started controllers ... ")

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
			<-sigCh

			log.Logger.Info("SIGTERM ... ")

			return nil
		},
	}
	return cmd
}
