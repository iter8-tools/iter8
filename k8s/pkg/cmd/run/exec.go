package run

import (
	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// complete sets all information needed for processing the command
func (o *Options) complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) (err error) {
	log.Logger.Trace("iter8 run complete() called")
	defer log.Logger.Trace("iter8 run complete() completed")

	o.namespace, _, err = factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.client, err = utils.GetClient(o.ConfigFlags)
	if err != nil {
		return err
	}

	if len(o.experimentId) == 0 {
		s, err := utils.GetExperimentSecret(o.client, o.namespace, o.experimentId)
		if err != nil {
			return err
		}
		o.experimentId = s.Labels[utils.IdLabel]
	}

	return err
}

// validate ensures that all required arguments and flag values are provided
func (o *Options) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *Options) run(cmd *cobra.Command, args []string) (err error) {
	log.Logger.Trace("iter8 run run() called")
	defer log.Logger.Trace("iter8 run run() completed")

	var expIO basecli.ExpIO
	if o.remote {
		expIO = &utils.KubernetesExpIO{
			Client:    o.client,
			Namespace: o.namespace,
			Name:      utils.SpecSecretPrefix + o.experimentId,
		}
	} else {
		expIO = &basecli.FileExpIO{}
	}

	log.Logger.Trace("iter8 run: build started")
	exp, err := basecli.Build(false, expIO)
	log.Logger.Trace("iter8 run: build finished")
	if err != nil {
		return err
	}

	return exp.Run(expIO)
}
