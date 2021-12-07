package report

import (
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
	expIO := &utils.KubernetesExpIO{
		Client:    o.client,
		Namespace: o.namespace,
		Name:      utils.SpecSecretPrefix + o.experimentId,
	}

	log.Logger.Trace("build started")
	exp, err := basecli.Build(true, expIO)
	log.Logger.Trace("build finished")
	if err != nil {
		return err
	}

	return exp.Report(basecli.ReportOptions.OutputFormat)
}
