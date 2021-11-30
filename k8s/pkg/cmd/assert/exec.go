package assert

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// complete sets all information needed for processing the command
func (o *Options) complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) (err error) {
	o.namespace, _, err = factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.client, err = utils.GetClient(o.ConfigFlags)
	if err != nil {
		return err
	}

	if len(o.experiment) == 0 {
		s, err := utils.GetExperiment(o.client, o.namespace, o.experiment)
		if err != nil {
			return err
		}
		o.experiment = s.GetName()
	}

	return err
}

// validate ensures that all required arguments and flag values are provided
func (o *Options) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *Options) run(cmd *cobra.Command, args []string) (err error) {
	var expIO basecli.ExpIO
	if !o.local {
		expIO = &utils.KubernetesExpIO{
			Client:    o.client,
			Namespace: o.namespace,
			Name:      o.experiment,
		}
	} else {
		expIO = &basecli.FileExpIO{}
	}

	log.Logger.Trace("build started")
	exp, err := basecli.Build(true, expIO)
	log.Logger.Trace("build finished")
	if err != nil {
		return err
	}

	allGood, err := exp.Assert(basecli.AssertOptions.Conds, basecli.AssertOptions.Timeout)
	if err != nil || !allGood {
		return err
	}

	if !allGood {
		return errors.New("")
	}

	return nil
}
